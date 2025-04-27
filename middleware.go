package hndlor

import (
	"net/http"
)

// NextHandler defines function signature for next handler
type NextHandler func(next http.Handler) http.Handler

// Middleware defines default function signature
type Middleware func(w http.ResponseWriter, r *http.Request, next http.Handler)

// MiddlewareWithError defines function signature
// with support for error capturing/response
type MiddlewareWithError func(w http.ResponseWriter, r *http.Request, next http.Handler) error

// M creates a middleware wrapper around the handler
func M(fn Middleware) NextHandler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r, next)
		})
	}
}

// MM creates a middleware wrapper around the handler
// with support for writing error response
func MM(fn MiddlewareWithError) NextHandler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := fn(w, r, next)
			if err != nil {
				_ = WriteError(w, err)
			}
		})
	}
}

// Chain accepts multiple [NextHandler] and builds new [NextHandler] as middleware
func Chain(mds ...NextHandler) NextHandler {
	return func(hnd http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mLen := len(mds)
			if mLen == 0 {
				hnd.ServeHTTP(w, r)
			} else {
				next := hnd
				for k := mLen - 1; k >= 0; k-- {
					next = mds[k](next)
				}
				next.ServeHTTP(w, r)
			}
		})
	}
}
