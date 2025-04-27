package hndlor_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/OpenRunic/hndlor"
)

func RunTestRequestBody(r http.Handler, method string, path string, body io.Reader, cbs ...func(*http.Request)) (*httptest.ResponseRecorder, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, fmt.Errorf("fail to create request: %s", err.Error())
	}

	if len(cbs) > 0 {
		cbs[0](req)
	}

	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	return res, nil
}

func RunTestJSONRequest(r http.Handler, method string, path string, data any) (*httptest.ResponseRecorder, error) {
	bodyBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return RunTestRequestBody(r, method, path, bytes.NewBuffer(bodyBytes), func(req *http.Request) {
		req.Header.Set("Content-Type", hndlor.ContentTypeJSON)
	})
}

func RunTestRequest(r http.Handler, method string, path string, cbs ...func(*http.Request)) (*httptest.ResponseRecorder, error) {
	return RunTestRequestBody(r, method, path, nil, cbs...)
}

func RunTestResultDecode(resp *http.Response, data any) error {
	defer resp.Body.Close()

	bt, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bt, data)
	if err != nil {
		return err
	}

	return nil
}

func InvalidateTestResultStatus(resp *http.Response, statusCode int) error {
	if resp.StatusCode != statusCode {
		return fmt.Errorf("invalid status code; expected %d but got %d", statusCode, resp.StatusCode)
	}
	return nil
}

func CreateTestRouter(paths ...string) *hndlor.MuxRouter {
	path := ""
	if len(paths) > 0 {
		path = paths[0]
	}
	return hndlor.SubRouter(path).Use(hndlor.PrepareMux())
}
