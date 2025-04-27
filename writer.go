package hndlor

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Json represents json data format
type Json map[string]any

// AsLoggable defines interface to check if struct is loggable
type AsLoggable interface {
	Log(io.Writer)
}

// WriteData writes [JsonData] to [io.Writer]
func WriteData(w io.Writer, data Json) error {
	bt, e := json.Marshal(data)
	if e != nil {
		return e
	}

	_, e = w.Write(bt)
	return e
}

// WriteError writes [error] to [io.Writer] and tries
// to use [AsJsonError] when available
func WriteError(w io.Writer, err error) error {
	var data Json
	statusCode := 0
	ex, ok := err.(AsExportableResponse)
	if ok {
		data = ex.ResponseJson()
		statusCode = ex.ResponseStatus()
	} else {
		data = Json{
			"error": err.Error(),
		}
	}

	if statusCode > 0 {
		rw, ok := w.(http.ResponseWriter)
		if ok {
			rw.WriteHeader(statusCode)
		}
	}

	LogError(log.Writer(), err)
	return WriteData(w, data)
}

// WriteError writes error message to [io.Writer]
func WriteErrorMessage(w io.Writer, err string) error {
	return WriteData(w, Json{
		"error": err,
	})
}

// LogError prints log to [io.Writer]
func LogError(w io.Writer, err error) {
	le, ok := err.(AsLoggable)
	if ok {
		le.Log(w)
	} else {
		fmt.Fprintf(w, "[ERR] %s\n", err.Error())
	}
}
