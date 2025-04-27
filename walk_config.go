package hndlor

import (
	"net/http"
	"os"
	"path/filepath"
)

// static indexes for [reflect.Field] positions
var wIndexes = []int{
	1, // ServeMux.routingNode [0]
	0, // routingNode.pattern [1]
	2, // routingNode.children [2]
	3, // routingNode.multiChild [3]
	4, // routingNode.emptyChild [4]
	0, // pattern.str [5]
	1, // pattern.method [6]
	3, // pattern.segments [7]
	4, // pattern.loc [8]
	0, // segment.s [9]
	1, // segment.wild [10]
	2, // segment.multi [11]
	0, // http.entry[string, *net/http.routingNode] [12]
	1, // mapping[string, *routingNode][1] [13]
	2, // pattern.host [14]
}

// working dir
var workingDir string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		workingDir = ""
	} else {
		workingDir = wd
	}
}

// resolves script path
func makeRouteLoc(path string) string {
	if len(workingDir) < 1 {
		return path
	}

	relPath, _ := filepath.Rel(workingDir, path)
	return relPath
}

// WalkConfig defines configurations for walk
type WalkConfig struct {
	Prefix   string
	Mappings map[string]*http.ServeMux
}

// Set registers mapping item
func (c *WalkConfig) Set(path string, mux *http.ServeMux) *WalkConfig {
	c.Mappings[path] = mux
	return c
}

// Has checks for mapping item
func (c *WalkConfig) Has(path string) bool {
	_, ok := c.Mappings[path]
	return ok
}

// Get retrieves mapping item
func (c *WalkConfig) Get(path string) *http.ServeMux {
	return c.Mappings[path]
}

// Clone creates new config with updated prefix
func (c *WalkConfig) Clone(prefix string) *WalkConfig {
	return &WalkConfig{
		Prefix:   prefix,
		Mappings: c.Mappings,
	}
}

// NewWalkConfig creates walk config instance
func NewWalkConfig() *WalkConfig {
	return &WalkConfig{
		Prefix:   "",
		Mappings: make(map[string]*http.ServeMux),
	}
}
