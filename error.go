package hndlor

import (
	"fmt"
	"io"
	"maps"
	"net/http"
)

// AsExportableResponse defines interface to export data for response
type AsExportableResponse interface {

	// ResponseStatus returns the http status code
	ResponseStatus() int

	// ResponseJSON exports the error as json data
	ResponseJSON() JSON
}

// ResponseError defines struct with error message and extras
type ResponseError struct {

	// main error message
	message string

	// path/location of error
	path string

	// reason behind the erro
	reason string

	// http status code
	statusCode int

	// error code on code level
	errorCode string

	// mark error as server error
	serverError bool

	// error message on server error
	clientMessage string

	// extra data to export
	extras JSON
}

// Status updates the status code for response
func (e *ResponseError) Status(c int) *ResponseError {
	e.statusCode = c
	return e
}

// ErrorCode updates primary error code
func (e *ResponseError) ErrorCode(c string) *ResponseError {
	e.errorCode = c
	return e
}

// Reason updates the reason of error
func (e *ResponseError) Reason(r string) *ResponseError {
	e.reason = r
	return e
}

// Path updates the request path info
func (e *ResponseError) Path(p string) *ResponseError {
	e.path = p
	return e
}

// Data updates the extra data for response
func (e *ResponseError) Extras(d JSON) *ResponseError {
	e.extras = d
	return e
}

// Server marks error as server error
func (e *ResponseError) Server() *ResponseError {
	e.serverError = true
	e.statusCode = http.StatusInternalServerError
	return e
}

// Client marks error as client side error
func (e *ResponseError) Client() *ResponseError {
	e.serverError = false
	return e
}

// Caller reads caller info to error
func (e *ResponseError) Caller(skip int) *ResponseError {
	file, line, ok := GetCaller(skip)
	if ok {
		return e.Path(fmt.Sprintf("%s:%d", file, line))
	}
	return e
}

// ClientMessage sets error message for exported server error
func (e *ResponseError) ClientMessage(m string) *ResponseError {
	e.clientMessage = m
	return e
}

func (e ResponseError) Log(w io.Writer) {
	if e.serverError {
		msg := e.message
		if len(e.reason) > 0 {
			msg = fmt.Sprintf("%s (%s)", msg, e.reason)
		}
		if len(e.path) > 0 {
			msg = fmt.Sprintf("[%s] %s", e.path, msg)
		}
		if len(e.errorCode) > 0 {
			msg = fmt.Sprintf("%s [%s]", msg, e.errorCode)
		}

		fmt.Fprintf(w, "Error: %s\n", msg)
	}
}

// Message returns underlying error message
func (e *ResponseError) Message() string {
	return e.message
}

// AsJSON generates json data for export
func (e *ResponseError) AsJSON() JSON {
	res := JSON{
		"error": e.message,
	}

	if len(e.errorCode) > 0 {
		res["code"] = e.errorCode
	}
	if len(e.reason) > 0 {
		res["reason"] = e.reason
	}

	if e.extras != nil {
		maps.Copy(res, e.extras)
	}

	return res
}

func (e ResponseError) ResponseStatus() int {
	return e.statusCode
}

func (e ResponseError) ResponseJSON() JSON {
	res := e.AsJSON()
	if e.serverError {
		if len(e.clientMessage) > 0 {
			res["error"] = e.clientMessage
		} else if e.statusCode > 0 {
			res["error"] = http.StatusText(e.statusCode)
		} else {
			res["error"] = "Unknown error"
		}
	}
	return res
}

func (e ResponseError) Error() string {
	if len(e.reason) > 0 {
		return fmt.Sprintf("%s (%s)", e.message, e.reason)
	}
	return e.message
}

// Error creates error with message only
func Error(msg string) *ResponseError {
	return &ResponseError{
		message: msg,
	}
}

// Errorf creates error with message and formatting options
func Errorf(format string, a ...any) *ResponseError {
	return Error(fmt.Sprintf(format, a...))
}
