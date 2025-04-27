package hndlor_test

import (
	"net/http"
	"testing"

	"github.com/OpenRunic/hndlor"
)

func TestMuxWalk(t *testing.T) {
	r := CreateTestRouter()
	r.HandleFunc("GET /a", func(w http.ResponseWriter, r *http.Request) {})
	r.HandleFunc("POST /b", func(w http.ResponseWriter, r *http.Request) {})
	r.HandleFunc("DELETE /c", func(w http.ResponseWriter, r *http.Request) {})

	dr := CreateTestRouter("/d")
	dr.HandleFunc("GET /d1", func(w http.ResponseWriter, r *http.Request) {})
	dr.HandleFunc("POST /d2", func(w http.ResponseWriter, r *http.Request) {})
	dr.HandleFunc("DELETE /d3", func(w http.ResponseWriter, r *http.Request) {})
	dr.MountTo(r)

	stats := hndlor.WalkCollect(r.Mux(), hndlor.NewWalkConfig().
		Set(dr.Path, dr.Mux()),
	)

	if len(stats) != 7 {
		t.Error("unable to walk through all routes")
	}
}
