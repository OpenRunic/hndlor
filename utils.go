package hndlor

import (
	"encoding/json"
	"net"
	"net/http"
	"runtime"
)

// RequestAddr retrieves the requesting address info
func RequestAddr(r *http.Request) net.Addr {
	return r.Context().Value(http.LocalAddrContextKey).(net.Addr)
}

// GetCaller retrieves the runtime.Caller info
func GetCaller(skip int) (string, int, bool) {
	_, file, no, ok := runtime.Caller(skip)
	return file, no, ok
}

// StructToStruct convert struct type
func StructToStruct(src any, data any) error {
	bt, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(bt, data)
}
