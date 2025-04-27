package hndlor

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ReadValue reads value as provided type or returns [error]
func ReadValue[T any](tp reflect.Type, value any, fb T) (T, error) {
	if tp == reflect.TypeOf(value) {
		return value.(T), nil
	}

	str := fmt.Sprint(value)
	switch tp.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fb, err
		}
		return any(val).(T), nil
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fb, err
		}
		return any(val).(T), nil
	case reflect.Bool:
		val, err := strconv.ParseBool(str)
		if err != nil {
			return fb, err
		}
		return any(val).(T), nil
	case reflect.String:
		return value.(T), nil
	}

	return fb, nil
}

// ReadFields retrieves all exported fields for the struct
func ReadFields(tp reflect.Type) []string {
	el := tp
	if el.Kind() == reflect.Ptr {
		el = el.Elem()
	}

	keys := make([]string, 0)
	fcount := el.NumField()
	for i := range fcount {
		f := el.Field(i)
		if f.IsExported() {
			keys = append(keys, strings.ToLower(f.Name))
		}
	}

	return keys
}
