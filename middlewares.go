package hndlor

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// lResponseWriter is a modified writer with logging info
type lResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *lResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Logger middleware builds handler to log every requests received
func Logger(lw io.Writer) NextHandler {
	return M(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		nw := &lResponseWriter{w, http.StatusOK}
		defer func(st time.Time) {
			fmt.Fprintf(lw, "[%s] %s - %s [%d]\n", r.Method, r.URL.Path, time.Since(st), nw.statusCode)
		}(time.Now())

		next.ServeHTTP(nw, r)
	})
}

// PrepareMux middleware parses request to create cache data as needed
func PrepareMux() NextHandler {
	return M(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		nr, err := PrepareBody(r)
		if err != nil {
			WriteError(w, Error(err.Error()).Server().Status(http.StatusUnprocessableEntity))
		} else {
			next.ServeHTTP(w, nr)
		}
	})
}
