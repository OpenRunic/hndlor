package hndlor

import (
	"net/http"
	"reflect"
)

type ValueFailHandler func(ValueResolver, error) error

// Handler defines struct for callback
type Handler struct {
	callback  any
	Err       error
	values    []ValueResolver
	valueFail ValueFailHandler
}

// OnFail defines callback when value resolve fails
func (h *Handler) OnFail(cb ValueFailHandler) *Handler {
	h.valueFail = cb
	return h
}

// Invalidate verifies the provided function with requested values
func (h *Handler) Invalidate() error {
	if h.Err == nil {
		vLen := len(h.values)
		tp := reflect.TypeOf(h.callback)

		if tp.Kind() != reflect.Func {
			h.Err = Errorf("invalid handler type; expected func got [ %s ]", tp).Server()
		} else {
			ins := make([]reflect.Type, vLen)
			for i := range vLen {
				ins[i] = h.values[i].Type()
			}
			outs := []reflect.Type{
				reflect.TypeOf(make(Json)),
				reflect.TypeOf((*error)(nil)).Elem(),
			}
			ep := reflect.FuncOf(ins, outs, false)

			if tp != ep {
				h.Err = Errorf("invalid handler function; expected [ %s ] got [ %s ]", ep, tp).Server()
			}
		}
	}

	return h.Err
}

// Values resolves the dynamic handler values
func (h *Handler) Values(r *http.Request) ([]reflect.Value, error) {
	vLen := len(h.values)
	values := make([]reflect.Value, vLen)

	for i := range vLen {
		value := h.values[i]
		val, err := value.Resolve(r)

		if err != nil {
			if h.valueFail != nil {
				err = h.valueFail(value, err)
			}

			if err != nil {
				return nil, err
			} else {
				values[i] = reflect.ValueOf(value.Default())
			}
		} else {
			values[i] = reflect.ValueOf(val)
		}
	}

	return values, nil
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if verr := h.Invalidate(); verr != nil {
		WriteError(w, verr)
		return
	}

	values, err := h.Values(r)
	if err != nil {
		WriteError(w, err)
		return
	}

	vt := reflect.ValueOf(h.callback)
	response := vt.Call(values)
	data, rerr := response[0].Interface(), response[1].Interface()

	if rerr != nil {
		WriteError(w, rerr.(error))
		return
	}

	WriteData(w, data.(Json))
}

// New creates [Handler] and panics on mis-matched function
// signature based on provided callback
//
// Example: reads query string 'name' and passes as func argument
//
//	mux.Handle("GET /hello", hndlor.New(func(name string) (hndlor.Json, error) {
//		return hndlor.Json{
//			"hello": name,
//		}, nil
//	}, hndlor.Get[string]("name")))
func New(cb any, values ...ValueResolver) *Handler {
	return &Handler{
		callback: cb,
		values:   values,
	}
}

// NewP works same as [New] but panics on invalidation error
func NewP(cb any, values ...ValueResolver) *Handler {
	h := New(cb, values...)

	if err := h.Invalidate(); err != nil {
		panic(err)
	}

	return h
}
