package hndlor

import (
	"context"
	"net/http"
)

// Values collects all provided values from *[http.Request]
func Values(r *http.Request, values ...ValueResolver) (JSON, error) {
	res := make(JSON)

	for _, val := range values {
		if len(val.Alias()) > 0 {
			v, err := val.Resolve(r)
			if err != nil {
				return nil, err
			}
			res[val.Alias()] = v
		}
	}

	return res, nil
}

// ValuesAs collects values and maps it to struct
func ValuesAs(r *http.Request, data any, values ...ValueResolver) error {
	vs, err := Values(r, values...)
	if err != nil {
		return err
	}

	return StructToStruct(vs, data)
}

// ReadRequest defines value resolver to access *[http.Request]
func ReadRequest() *Value[*http.Request] {
	return NewValue[*http.Request]("", ValueSourceDefault).
		Reader(func(r *http.Request) (*http.Request, error) {
			return r, nil
		})
}

// ReadContext defines value resolver to access request's [context.Context]
func ReadContext() *Value[context.Context] {
	return NewValue[context.Context]("", ValueSourceDefault).
		Reader(func(r *http.Request) (context.Context, error) {
			return r.Context(), nil
		})
}
