package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type httpResponse struct {
	Code int
	Body []byte
}

func doRequest(t *testing.T, h http.Handler, method, path string, body any, headers map[string]string) httpResponse {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return httpResponse{Code: rec.Code, Body: rec.Body.Bytes()}
}

func extractField[T any](t *testing.T, data []byte, field string) T {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	v, ok := m[field]
	if !ok {
		t.Fatalf("missing field %s", field)
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	var out T
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("cast field %s: %v", field, err)
	}
	return out
}

func TestTopDown_Flow_AdminAuthUsersTasks(t *testing.T) {
	// top-down: start the whole app via router and hit HTTP endpoints
	r := BuildRouter()

	// 1) Login as seeded admin
	loginBody := map[string]any{"username": "admin", "password": "admin123"}
	resp := doRequest(t, r, http.MethodPost, "/api/auth/login", loginBody, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("login status = %d, body=%s", resp.Code, string(resp.Body))
	}
	accessToken := extractField[string](t, resp.Body, "accessToken")
	if accessToken == "" {
		t.Fatalf("expected accessToken")
	}

	authz := map[string]string{"Authorization": "Bearer " + accessToken}

	// 2) Admin: list users
	resp = doRequest(t, r, http.MethodGet, "/api/users", nil, authz)
	if resp.Code != http.StatusOK {
		t.Fatalf("users list status = %d, body=%s", resp.Code, string(resp.Body))
	}

	// 3) Create a task (no auth needed for tasks per current design)
	createTask := map[string]any{"title": "Write tests", "description": "top-down"}
	resp = doRequest(t, r, http.MethodPost, "/tasks", createTask, nil)
	if resp.Code != http.StatusCreated {
		t.Fatalf("create task status = %d, body=%s", resp.Code, string(resp.Body))
	}
	createdID := extractField[float64](t, resp.Body, "id")
	if int(createdID) <= 0 {
		t.Fatalf("expected id > 0")
	}

	// 4) Get tasks list
	resp = doRequest(t, r, http.MethodGet, "/tasks", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("list tasks status = %d, body=%s", resp.Code, string(resp.Body))
	}
}
