package hndlor

import (
	"context"
	"maps"
	"net/http"
)

// ContextValue defines custom type for [context.Context] key
type ContextValue int

const (
	ContextValueDefault ContextValue = iota // default context key for data
	ContextValueJSON
)

// GetAllData retrieves saved [JSON] saved in default context data
func GetAllData(r *http.Request) JSON {
	var data JSON
	raw := r.Context().Value(ContextValueDefault)
	if raw == nil {
		data = JSON{}
	} else {
		data = raw.(JSON)
	}

	return data
}

// GetData retrieves specific key from saved default context data
func GetData[T any](r *http.Request, key string, fb T) (T, error) {
	data := GetAllData(r)
	v, ok := data[key]
	if ok {
		return v.(T), nil
	}
	return fb, Errorf("unable to find context data: %s", key).Server().Path(r.URL.Path)
}

// PatchValue writes key/value to default context data
func PatchValue(r *http.Request, key string, value any) *http.Request {
	data := GetAllData(r)
	data[key] = value

	return Patch(r, ContextValueDefault, data)
}

// PatchValue writes [JSON] to default context data
func PatchMap(r *http.Request, value JSON) *http.Request {
	data := GetAllData(r)
	maps.Copy(data, value)

	return Patch(r, ContextValueDefault, data)
}

// Patch updates the request context with new key/value data
func Patch(r *http.Request, key any, value any) *http.Request {
	return r.WithContext(
		context.WithValue(r.Context(), key, value),
	)
}
