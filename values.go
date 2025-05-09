package hndlor

import (
	"context"
	"net/http"
)

// Values collects all provided values from *[http.Request]
func Values(w http.ResponseWriter, r *http.Request, values ...ValueResolver) (JSON, error) {
	res := make(JSON)

	for _, val := range values {
		if len(val.Alias()) > 0 {
			v, err := val.Resolve(w, r)
			if err != nil {
				return nil, err
			}
			res[val.Alias()] = v
		}
	}

	return res, nil
}

// ValuesAs collects values and maps it to struct
func ValuesAs(w http.ResponseWriter, r *http.Request, data any, values ...ValueResolver) error {
	vs, err := Values(w, r, values...)
	if err != nil {
		return err
	}

	return StructToStruct(vs, data)
}

// HTTPRequest defines value resolver to access *[http.Request]
func HTTPRequest() *Value[*http.Request] {
	return NewValue[*http.Request]("", ValueSourceDefault).
		Reader(func(_ http.ResponseWriter, r *http.Request) (*http.Request, error) {
			return r, nil
		})
}

// HTTPResponseWriter defines value resolver to access [http.ResponseWriter]
func HTTPResponseWriter() *Value[http.ResponseWriter] {
	return NewValue[http.ResponseWriter]("", ValueSourceDefault).
		Reader(func(w http.ResponseWriter, _ *http.Request) (http.ResponseWriter, error) {
			return w, nil
		})
}

// HTTPContext defines value resolver to access request's [context.Context]
func HTTPContext() *Value[context.Context] {
	return NewValue[context.Context]("", ValueSourceDefault).
		Reader(func(_ http.ResponseWriter, r *http.Request) (context.Context, error) {
			return r.Context(), nil
		})
}
