package hndlor

import "net/http"

// MountableMux defines an interface to verify if it can handle requests
type MountableMux interface {
	Handle(string, http.Handler)
}

// MuxRouter defines a helper router
type MuxRouter struct {
	Path        string
	Middlewares []NextHandler
	mux         *http.ServeMux
}

// Get internal mux router *[http.ServeMux]
func (g *MuxRouter) Mux() *http.ServeMux {
	return g.mux
}

// Use adds [NextHandler] as middlewares for all routes
func (g *MuxRouter) Use(hns ...NextHandler) *MuxRouter {
	g.Middlewares = append(g.Middlewares, hns...)
	return g
}

// MountTo attaches [MuxRouter] to parent [MountableMux]
func (g MuxRouter) MountTo(target MountableMux) {
	if len(g.Path) > 0 {
		target.Handle(g.Path+"/", http.StripPrefix(g.Path, g))
	}
}

// Handle adds new request handler [http.Handler]
func (g *MuxRouter) Handle(pattern string, handler http.Handler) {
	g.mux.Handle(pattern, handler)
}

// HandleFunc adds new request handler func [http.HandlerFunc]
func (g *MuxRouter) HandleFunc(pattern string, handler http.HandlerFunc) {
	g.mux.HandleFunc(pattern, handler)
}

// ServerHTTP server the response
func (g MuxRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Chain(g.Middlewares...)(g.mux).ServeHTTP(w, r)
}

// SubRouter creates new router instance with path
func SubRouter(path string) *MuxRouter {
	return &MuxRouter{
		Path:        path,
		mux:         http.NewServeMux(),
		Middlewares: make([]NextHandler, 0),
	}
}

// Router creates new parent router instance
func Router() *MuxRouter {
	return SubRouter("")
}
