package hndlor

import (
	"net/http"
	"reflect"
)

// ValueSource defines the source of value
type ValueSource int

const (
	ValueSourcePath    ValueSource = iota // reads from path params
	ValueSourceGet                        // reads from url query
	ValueSourceBody                       // reads from request body
	ValueSourceHeader                     // reads from request header
	ValueSourceContext                    // reads from request context default data
	ValueSourceDefault                    // reads from source based on request method
)

// ValueResolver defines an interface to be used by handler
type ValueResolver interface {

	// Field returns the name of the field
	Field() string

	// Alias returns the alias of the field
	Alias() string

	// Type returns [reflect.Type] of value
	Type() reflect.Type

	// Default returns default value of type
	Default() any

	// Checks if value is required
	Required() bool

	// Resolve evaluates and return the resulting value else error
	Resolve(*http.Request) (any, error)
}

// Value[T any] defines struct for resolving value from sources
type Value[T any] struct {

	// name of the field
	field string

	// alias name of the field
	alias string

	// if value is required
	required bool

	// source of value to read from
	source ValueSource

	// custom value reader
	reader func(*http.Request) (T, error)

	// validate value
	validate func(*http.Request, T) error

	// [reflect.Type] resolved from [T]
	rType reflect.Type

	// default value resolved for given [T]
	rDefault T
}

func (v *Value[T]) Field() string {
	return v.field
}

func (v *Value[T]) Alias() string {
	if len(v.alias) > 0 {
		return v.alias
	}
	return v.field
}

func (v *Value[T]) Type() reflect.Type {
	return v.rType
}

func (v *Value[T]) Default() any {
	return v.rDefault
}

func (v *Value[T]) Required() bool {
	return v.required
}

// As sets alias name for field
func (v *Value[T]) As(n string) *Value[T] {
	v.alias = n
	return v
}

// Optional marks value resolver as optional value
func (v *Value[T]) Optional() *Value[T] {
	v.required = false
	return v
}

// Validate adds value validator to resolved value
func (v *Value[T]) Validate(cb func(*http.Request, T) error) *Value[T] {
	v.validate = cb
	return v
}

// Reader stores custom value reader for value resolver
func (v *Value[T]) Reader(cb func(*http.Request) (T, error)) *Value[T] {
	v.reader = cb
	return v
}

// readValue reads the value from *[http.Request] for provided [ValueSource]
func (v *Value[T]) readValue(r *http.Request, src ValueSource) (T, error) {
	asStruct := (v.rType.Kind() == reflect.Struct ||
		(v.rType.Kind() == reflect.Ptr && v.rType.Elem().Kind() == reflect.Struct))

	if asStruct {
		var data T

		if src == ValueSourceBody {
			err := BodyReadStruct(r, &data)
			if err != nil {
				return v.rDefault, err
			}
			return data, nil
		}

		fields := ReadFields(v.rType)
		values := make(map[string]any)

		switch src {
		case ValueSourceGet:
			for key := range r.URL.Query() {
				values[key] = r.URL.Query().Get(key)
			}
		case ValueSourcePath:
			for _, key := range fields {
				if len(r.PathValue(key)) > 0 {
					values[key] = r.PathValue(key)
				}
			}
		case ValueSourceHeader:
			for _, key := range fields {
				if len(r.Header.Get(key)) > 0 {
					values[key] = r.Header.Get(key)
				}
			}
		case ValueSourceContext:
			for _, key := range fields {
				kv, err := GetData[any](r, key, nil)
				if err != nil {
					return v.rDefault, err
				}
				values[key] = kv
			}
		}

		if len(values) > 0 {
			err := StructToStruct(values, &data)
			if err != nil {
				return v.rDefault, err
			}
			return data, nil
		}
	} else {
		switch src {
		case ValueSourceGet:
			if r.URL != nil && r.URL.Query().Has(v.field) {
				return ReadValue(v.rType, r.URL.Query().Get(v.field), v.rDefault)
			}
		case ValueSourceBody:
			bValue, ok := BodyRead(r, v.field)
			if ok {
				return ReadValue(v.rType, bValue, v.rDefault)
			}
		case ValueSourcePath:
			pVal := r.PathValue(v.field)
			if len(pVal) > 0 {
				return ReadValue(v.rType, pVal, v.rDefault)
			}
		case ValueSourceHeader:
			hVal, ok := r.Header[v.field]
			if ok {
				return ReadValue(v.rType, hVal[0], v.rDefault)
			}
		case ValueSourceContext:
			return GetData(r, v.field, v.rDefault)
		}
	}

	return v.rDefault, Errorf("resolve value failed [%s]", v.field).Reason("value_failed")
}

// readDefaultValue reads the value from *[http.Request] based on request method
func (v Value[T]) readDefaultValue(r *http.Request) (T, error) {
	var value T
	var err error

	sources := make([]ValueSource, 0)
	if HasBody(r) {
		sources = append(sources, ValueSourceBody)
	} else {
		sources = append(sources, ValueSourceGet, ValueSourcePath)
	}

	for _, t := range sources {
		value, err = v.readValue(r, t)
		if err == nil {
			return value, nil
		}
	}

	return value, err
}

func (v Value[T]) Resolve(r *http.Request) (any, error) {
	if v.reader != nil {
		return v.reader(r)
	}

	var value T
	var err error
	if v.source == ValueSourceDefault {
		value, err = v.readDefaultValue(r)
	} else {
		value, err = v.readValue(r, v.source)
	}

	if err != nil {
		if v.required {
			return v.rDefault, err
		}

		return v.rDefault, nil
	}

	if v.validate != nil {
		err := v.validate(r, value)
		if err != nil {
			return v.rDefault, err
		}
	}

	return value, nil
}

// NewValue define new value resolver instance
func NewValue[T any](field string, src ValueSource) *Value[T] {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	return &Value[T]{
		field:    field,
		rType:    tp,
		rDefault: (reflect.New(tp).Elem().Interface()).(T),
		source:   src,
		required: true,
	}
}

// StructFrom defines new struct resolver instance
func StructFrom[T any](src ValueSource) *Value[T] {
	return NewValue[T]("", src)
}

// Struct defines struct resolver for default source
func Struct[T any]() *Value[T] {
	return StructFrom[T](ValueSourceDefault)
}

// Get defines value resolver from query string
func Get[T any](field string) *Value[T] {
	return NewValue[T](field, ValueSourceGet)
}

// Body defines value resolver from request body
func Body[T any](field string) *Value[T] {
	return NewValue[T](field, ValueSourceBody)
}

// Path defines value resolver from url path params
func Path[T any](field string) *Value[T] {
	return NewValue[T](field, ValueSourcePath)
}

// Header defines value resolver from request header
func Header[T any](field string) *Value[T] {
	return NewValue[T](field, ValueSourceHeader)
}

// Context defines value resolver from default data on request context
func Context[T any](key string) *Value[T] {
	return NewValue[T](key, ValueSourceContext)
}

// Reader defines value resolver using custom reader
func Reader[T any](cb func(*http.Request) (T, error)) *Value[T] {
	return NewValue[T]("", ValueSourceDefault).Reader(cb)
}

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
