package hndlor

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
)

// WalkCallback defines function signature for walk callback
type WalkCallback func(RouteStat)

// RouteStat defines struct for collected route info
type RouteStat struct {
	Str    string
	Path   string
	Method string
	Host   string
	Loc    string
	Group  bool
	Prefix string
}

func (s RouteStat) String() string {
	m := s.Method
	if len(m) < 1 {
		m = "GET"
	}
	if s.Group {
		m = "MUX"
	}

	host := ""
	if len(s.Host) > 0 {
		host = fmt.Sprintf("(%s) ", s.Host)
	}

	return fmt.Sprintf("[%s] %s%s%s - %s", m, host, s.Prefix, s.Path, s.Loc)
}

// buildRouteStat generates [RouteStat] from [reflect.Value] of routingNode
func buildRouteStat(prefix string, node reflect.Value) RouteStat {
	ec := node.Field(wIndexes[4])
	if !ec.IsNil() {
		node = ec.Elem()
	}

	pattern := node.Field(wIndexes[1]).Elem()
	segments := pattern.Field(wIndexes[7])

	grp := false
	n := segments.Len()
	parts := make([]string, n)
	for i := range n {
		p := segments.Index(i)
		part := p.Field(wIndexes[9]).String()

		if len(part) > 0 {
			if p.Field(wIndexes[10]).Bool() {
				if p.Field(wIndexes[11]).Bool() {
					part = fmt.Sprintf("{%s...}", part)
				} else {
					part = fmt.Sprintf("{%s}", part)
				}
			}
			parts[i] = part
		} else if p.Field(wIndexes[10]).Bool() {
			grp = true
		}
	}

	return RouteStat{
		Group:  grp,
		Prefix: prefix,
		Str:    pattern.Field(wIndexes[5]).String(),
		Path:   "/" + strings.Join(parts, "/"),
		Method: pattern.Field(wIndexes[6]).String(),
		Host:   pattern.Field(wIndexes[14]).String(),
		Loc:    makeRouteLoc(pattern.Field(wIndexes[8]).String()),
	}
}

// walkNode parses [reflect.Value] of routingNode
// to collect all available routes
func walkNode(config *WalkConfig, node reflect.Value, cb WalkCallback) {
	if node.Type().Name() != "routingNode" {
		return
	}

	mc := node.Field(wIndexes[3])
	ec := node.Field(wIndexes[4])
	ch := node.Field(wIndexes[2]).Field(wIndexes[12])

	ex := false
	if !ec.IsNil() {
		ex = true
		walkNode(config, ec.Elem(), cb)
	}
	if !mc.IsNil() {
		ex = true
		walkNode(config, mc.Elem(), cb)
	}

	if !ch.IsZero() {
		walkChildren(config, ch, cb)
	} else if !ex {
		stat := buildRouteStat(config.Prefix, node)
		cb(stat)

		path := strings.TrimRight(stat.Path, "/")
		if stat.Group && config.Has(path) {
			Walk(config.Get(path), cb, config.Clone(path))
		}
	}
}

// walkChildren parses [reflect.Value] of child
// routes to collect all available routes
func walkChildren(config *WalkConfig, nodes reflect.Value, cb WalkCallback) {
	if nodes.Type().Kind() != reflect.Slice {
		return
	}

	for i := range nodes.Len() {
		walkNode(config, nodes.Index(i).Field(wIndexes[13]).Elem(), cb)
	}
}

// Walk parses the *[http.ServeMux] using reflection
// to collect list of routes via callback [WalkCallback]
//
// Route collection from nested *[http.ServeMux] isn't
// supported unless details are provided via [WalkConfig]
func Walk(mux *http.ServeMux, cb WalkCallback, configs ...*WalkConfig) {
	var config *WalkConfig
	if len(configs) > 0 {
		config = configs[0]
	} else {
		config = NewWalkConfig()
	}

	tp := reflect.ValueOf(mux).Elem()
	walkNode(config, tp.Field(wIndexes[0]), cb)
}

// WalkCollect walks through mux to generate route stats
func WalkCollect(mux *http.ServeMux, configs ...*WalkConfig) []RouteStat {
	stats := make([]RouteStat, 0)

	Walk(mux, func(rs RouteStat) {
		stats = append(stats, rs)
	}, configs...)

	return stats
}

// WriteStats logs info of defined routes to [io.Writer]
func WriteStats(mux *http.ServeMux, w io.Writer, configs ...*WalkConfig) {
	fmt.Fprint(w, "\n==================================\n")
	fmt.Fprint(w, "            Routes list")
	fmt.Fprint(w, "\n==================================\n\n")

	Walk(mux, func(rs RouteStat) {
		fmt.Fprintf(w, "%s\n", rs)
	}, configs...)

	fmt.Fprint(w, "\n==================================\n\n")
}
