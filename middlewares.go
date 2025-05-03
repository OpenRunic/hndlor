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
	contentSize uint64
	statusCode  int
}

func (w *lResponseWriter) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	if err == nil {
		w.contentSize = uint64(size)
	} else {
		w.contentSize = 0
	}
	return size, err
}

func (w *lResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Logger middleware builds handler to log every requests received
func Logger(lw io.Writer) NextHandler {
	return M(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		nw := &lResponseWriter{w, 0, http.StatusOK}
		defer func(st time.Time) {
			fmt.Fprintf(lw, "[%s] %s - (T %s, S %d, L %d)\n",
				r.Method,
				r.URL.Path,
				time.Since(st),
				nw.statusCode,
				nw.contentSize,
			)
		}(time.Now())

		next.ServeHTTP(nw, r)
	})
}

// PrepareMux middleware parses request to create cache data as needed
func PrepareMux() NextHandler {
	return M(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		nr, err := PrepareBody(r)
		if err != nil {
			_ = WriteError(w, Error(err.Error()).Server().Status(http.StatusUnprocessableEntity))
		} else {
			next.ServeHTTP(w, nr)
		}
	})
}
