package hndlor_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/OpenRunic/hndlor"
)

type TestLoginCredentials struct {
	Username string
	Password string
}

func CreateMethodTestRouter() *hndlor.MuxRouter {
	r := CreateTestRouter()

	r.Handle("GET /hello/{name}", hndlor.New(func(name string) (hndlor.JSON, error) {
		return hndlor.JSON{
			"message": fmt.Sprintf("Hello %s!", name),
		}, nil
	}, hndlor.Path[string]("name")))

	authGroup := CreateTestRouter("/auth")
	authGroup.Handle("POST /login", hndlor.New(func(creds TestLoginCredentials) (hndlor.JSON, error) {
		return hndlor.JSON{
			"username": creds.Username,
			"password": creds.Password,
		}, nil
	}, hndlor.Struct[TestLoginCredentials]()))
	authGroup.MountTo(r)

	return r
}

func TestGetRoute(t *testing.T) {
	r := CreateMethodTestRouter()

	res, err := RunTestRequest(r, "GET", "/hello/John")
	if err != nil {
		t.Fatal(err)
	}
	response := res.Result()

	err = InvalidateTestResultStatus(response, 200)
	if err != nil {
		t.Error(err)
	} else {
		var data hndlor.JSON
		err := RunTestResultDecode(response, &data)
		if err != nil {
			t.Error(err)
		} else if data["message"] != "Hello John!" {
			t.Error("unable to resolve valid response data on GET")
		}
	}
}

func TestPostRoute(t *testing.T) {
	r := CreateMethodTestRouter()
	sampleLoginData := TestLoginCredentials{
		Username: "admin",
		Password: "pass",
	}

	res, err := RunTestJSONRequest(r, "POST", "/auth/login", sampleLoginData)
	if err != nil {
		t.Fatal(err)
	}
	response := res.Result()

	err = InvalidateTestResultStatus(response, 200)
	if err != nil {
		t.Error(err)
	} else {
		var data TestLoginCredentials
		err := RunTestResultDecode(response, &data)
		if err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(sampleLoginData, data) {
			t.Error("unable to resolve valid response data on POST")
		}
	}
}
