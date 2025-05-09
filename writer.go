package hndlor

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// JSON represents json data format
type JSON map[string]any

// AsLoggable defines interface to check if struct is loggable
type AsLoggable interface {
	Log(io.Writer)
}

// WriteData writes [JSON] to [io.Writer]
func WriteData(w io.Writer, data JSON) error {
	bt, e := json.Marshal(data)
	if e != nil {
		return e
	}

	rs, ok := w.(http.ResponseWriter)
	if ok {
		rs.Header().Add("Content-Type", "application/json")
	}

	_, e = w.Write(bt)
	return e
}

// WriteError writes [error] to [io.Writer] and tries
// to use [AsExportableResponse] when available
func WriteError(w io.Writer, err error) error {
	var data JSON
	statusCode := 0
	ex, ok := err.(AsExportableResponse)
	if ok {
		data = ex.ResponseJSON()
		statusCode = ex.ResponseStatus()
	} else {
		data = JSON{
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

// WriteMessage writes message to [io.Writer]
func WriteMessage(w io.Writer, msg string) error {
	return WriteData(w, JSON{
		"message": msg,
	})
}

// WriteError writes error message to [io.Writer]
func WriteErrorMessage(w io.Writer, err string) error {
	return WriteData(w, JSON{
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
