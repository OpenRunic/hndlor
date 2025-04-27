package hndlor_test

import (
	"net/http"
	"testing"

	"github.com/OpenRunic/hndlor"
)

func CreateMiddlewareTestRouter() *hndlor.MuxRouter {
	r := CreateTestRouter()
	r.Use(hndlor.MM(func(w http.ResponseWriter, r *http.Request, next http.Handler) error {
		token := r.Header.Get("x-api-token")
		if len(token) < 1 {
			return hndlor.Error("auth token missing").Status(http.StatusForbidden)
		}

		next.ServeHTTP(w, hndlor.PatchValue(r, "authToken", token))
		return nil
	}))
	r.Handle("GET /me", hndlor.New(func(token string) (hndlor.Json, error) {
		return hndlor.Json{
			"token": token,
		}, nil
	}, hndlor.Context[string]("authToken")))

	return r
}

func TestRouteMiddleware(t *testing.T) {
	r := CreateMiddlewareTestRouter()

	res, err := RunTestRequest(r, "GET", "/me", func(r *http.Request) {
		r.Header.Set("x-api-token", "111")
	})
	if err != nil {
		t.Fatal(err)
	}
	response := res.Result()

	err = InvalidateTestResultStatus(response, 200)
	if err != nil {
		t.Error(err)
	} else {
		var data hndlor.Json
		err := RunTestResultDecode(response, &data)
		if err != nil {
			t.Error(err)
		} else if data["token"] != "111" {
			t.Error("unable to resolve valid token on response data")
		}
	}
}

func TestRouteMiddlewareFail(t *testing.T) {
	r := CreateMiddlewareTestRouter()

	res, err := RunTestRequest(r, "GET", "/me")
	if err != nil {
		t.Fatal(err)
	}
	response := res.Result()

	err = InvalidateTestResultStatus(response, 403)
	if err != nil {
		t.Error(err)
	}
}
