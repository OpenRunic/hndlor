package hndlor

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"slices"
	"strings"
)

// ContentType of json
const ContentTypeJson = "application/json"

// ContentType of url encoded form
const ContentTypeUrlEncoded = "application/x-www-form-urlencoded"

// ContentType of multipart
const ContentTypeMultipart = "multipart/form-data"

// HasBody checks if request has body
func HasBody(r *http.Request) bool {
	return slices.Contains([]string{"POST", "PUT", "PATCH"}, r.Method)
}

// PrepareBody parses any body request
func PrepareBody(r *http.Request) (*http.Request, error) {
	if HasBody(r) {
		cType := r.Header.Get("Content-Type")

		switch {
		case strings.Contains(cType, ContentTypeMultipart):
			err := r.ParseMultipartForm(0)
			if err != nil {
				return nil, err
			}

		case cType == ContentTypeUrlEncoded:
			err := r.ParseForm()
			if err != nil {
				return nil, err
			}

		case cType == ContentTypeJson, cType == "":
			eJson := BodyJson(r)
			if eJson != nil {
				return r, nil
			}

			var data Json
			err := json.NewDecoder(r.Body).Decode(&data)
			if err != nil {
				return nil, err
			}
			return Patch(r, ContextValueJson, data), nil
		}
	}

	return r, nil
}

// BodyJson reads the loaded json data from request context
func BodyJson(r *http.Request) Json {
	raw := r.Context().Value(ContextValueJson)
	if raw != nil {
		return raw.(Json)
	}
	return nil
}

// BodyRead reads value from request body
func BodyRead(r *http.Request, key string) (any, bool) {
	if r.Form != nil && r.Form.Has(key) {
		return r.FormValue(key), true
	} else if r.PostForm != nil && r.PostForm.Has(key) {
		return r.PostFormValue(key), true
	} else {
		data := BodyJson(r)
		if data != nil {
			v, ok := data[key]
			return v, ok
		}
	}
	return "", false
}

// BodyReadStruct reads values from request body as struct
func BodyReadStruct[T any](r *http.Request, data T) error {
	err := errors.New("failed to decode body")

	if HasBody(r) {
		cType := r.Header.Get("Content-Type")
		if cType == ContentTypeJson {
			jData := BodyJson(r)
			if jData == nil {
				return err
			}

			return StructToStruct(jData, data)
		}

		fields := ReadFields(reflect.TypeOf(data))
		values := make(map[string]any)
		for _, key := range fields {
			if r.Form != nil && r.Form.Has(key) {
				values[key] = r.FormValue(key)
			} else if r.PostForm != nil && r.PostForm.Has(key) {
				values[key] = r.PostFormValue(key)
			}
		}

		if len(values) > 0 {
			return StructToStruct(values, data)
		}
	}

	return err
}
