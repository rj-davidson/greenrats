package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/rj-davidson/greenrats/ent"
)

type TestApp struct {
	t   *testing.T
	App *fiber.App
}

func NewTestApp(t *testing.T) *TestApp {
	t.Helper()
	return &TestApp{
		t:   t,
		App: fiber.New(),
	}
}

func (ta *TestApp) WithAuth(ac *AuthContext) *TestApp {
	ta.App.Use(InjectAuth(ac))
	return ta
}

func (ta *TestApp) WithDBUser(user *ent.User) *TestApp {
	ac := DefaultAuthContext().WithDBUser(user)
	ta.App.Use(InjectAuth(ac))
	return ta
}

type TestResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
}

func (tr *TestResponse) JSON(v any) error {
	return json.Unmarshal(tr.Body, v)
}

func (tr *TestResponse) String() string {
	return string(tr.Body)
}

func (ta *TestApp) Get(path string) *TestResponse {
	return ta.doRequest("GET", path, nil)
}

func (ta *TestApp) Post(path string, body any) *TestResponse {
	return ta.doRequest("POST", path, body)
}

func (ta *TestApp) Put(path string, body any) *TestResponse {
	return ta.doRequest("PUT", path, body)
}

func (ta *TestApp) Patch(path string, body any) *TestResponse {
	return ta.doRequest("PATCH", path, body)
}

func (ta *TestApp) Delete(path string) *TestResponse {
	return ta.doRequest("DELETE", path, nil)
}

func (ta *TestApp) doRequest(method, path string, body any) *TestResponse {
	ta.t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			ta.t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := ta.App.Test(req)
	if err != nil {
		ta.t.Fatalf("failed to execute test request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		ta.t.Fatalf("failed to read response body: %v", err)
	}

	headers := make(map[string][]string)
	for k, v := range resp.Header {
		headers[k] = v
	}

	return &TestResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    headers,
	}
}

func (ta *TestApp) DoRequest(method, path string, body any, headers map[string]string) *TestResponse {
	ta.t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			ta.t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := ta.App.Test(req)
	if err != nil {
		ta.t.Fatalf("failed to execute test request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		ta.t.Fatalf("failed to read response body: %v", err)
	}

	respHeaders := make(map[string][]string)
	for k, v := range resp.Header {
		respHeaders[k] = v
	}

	return &TestResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    respHeaders,
	}
}
