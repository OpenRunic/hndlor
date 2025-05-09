package hndlor_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/OpenRunic/hndlor"
)

func TestValueResolve(t *testing.T) {
	body := hndlor.JSON{
		"username": "admin",
		"password": "pass",
	}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "/?name=John", bytes.NewBuffer(bodyBytes))

	if err != nil {
		t.Error(err)
	} else {
		req.SetPathValue("uid", "100")
		req.Header.Set("Content-Type", hndlor.ContentTypeJSON)
		req.Header.Set("x-api-token", "xyz")
		req, _ = hndlor.PrepareBody(hndlor.PatchValue(req, "identifier", "sample-iden"))

		_, err := hndlor.Values(nil, req,
			hndlor.Get[string]("name"),
			hndlor.Get[string]("q").Optional(),
			hndlor.Body[string]("username"),
			hndlor.Path[string]("uid"),
			hndlor.Header[string]("X-Api-Token").As("token"),
			hndlor.Context[string]("identifier"),
			hndlor.Reader(func(_ http.ResponseWriter, _ *http.Request) (int, error) {
				return 10, nil
			}).As("rank"),
			hndlor.Struct[TestLoginCredentials]().As("login").Validate(func(_ *http.Request, tlc TestLoginCredentials) error {
				if len(tlc.Username) > 0 {
					return nil
				}
				return errors.New("unable to resolve login data")
			}),
			hndlor.HTTPRequest().As("req"),
			hndlor.HTTPContext().As("context"),
		)
		if err != nil {
			t.Error(err)
		}
	}
}
