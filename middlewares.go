package hndlor

import (
	"fmt"
	"io"
	"log/slog"
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
func Logger(lw any) NextHandler {
	var ok bool
	var target string
	var writer io.Writer
	var sLogger *slog.Logger

	sLogger, ok = lw.(*slog.Logger)
	if ok {
		target = "slog"
	} else {
		writer, ok = lw.(io.Writer)
		if ok {
			target = "writer"
		}
	}

	return M(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		nw := &lResponseWriter{w, 0, http.StatusOK}

		if len(target) > 0 {
			defer func(st time.Time) {
				etime := time.Since(st)

				switch target {
				case "slog":
					sLogger.Info(
						"http request",
						"method", r.Method,
						"path", r.URL.Path,
						"time_ms", etime,
						"status", nw.statusCode,
						"size", nw.contentSize,
					)
				case "writer":
					fmt.Fprintf(writer, "[%s] %s - (T %s, S %d, L %d)\n",
						r.Method,
						r.URL.Path,
						etime,
						nw.statusCode,
						nw.contentSize,
					)
				}
			}(time.Now())
		}

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
