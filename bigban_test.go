package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

const baseURL = "http://localhost:8080"

var accessToken string

// вспомогательная функция
func doRequest(t *testing.T, method, url string, body any, auth bool) *http.Response {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	}
	req, err := http.NewRequest(method, baseURL+url, reqBody)
	if err != nil {
		t.Fatalf("ошибка создания запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if auth && accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("ошибка выполнения запроса %s %s: %v", method, url, err)
	}
	return resp
}

// Тест-цепочка «Большой взрыв»
func TestBigBang(t *testing.T) {
	// 1. Логин админа
	loginBody := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	resp := doRequest(t, "POST", "/api/auth/login", loginBody, false)
	if resp.StatusCode != 200 {
		t.Fatalf("ожидался 200 при логине, получили %d", resp.StatusCode)
	}
	var loginResp map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&loginResp)
	resp.Body.Close()
	if tok, ok := loginResp["accessToken"].(string); ok {
		accessToken = tok
	} else {
		t.Fatal("не удалось извлечь accessToken")
	}

	// 2. Получение списка пользователей
	resp = doRequest(t, "GET", "/api/users", nil, true)
	if resp.StatusCode != 200 {
		t.Errorf("GET /users: ожидался 200, получили %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 3. Создание задачи
	taskBody := map[string]any{
		"title":       "Написать отчёт",
		"description": "Метод Big Bang",
		"completed":   false,
	}
	resp = doRequest(t, "POST", "/tasks", taskBody, true)
	if resp.StatusCode != 201 {
		t.Errorf("POST /tasks: ожидался 201, получили %d", resp.StatusCode)
	}
	var taskResp map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&taskResp)
	resp.Body.Close()
	taskID, ok := taskResp["id"].(float64)
	if !ok {
		t.Fatal("не удалось извлечь id задачи")
	}

	// 4. Обновление задачи
	updateBody := map[string]any{
		"title":       "Отчёт",
		"description": "Big Bang завершён",
		"completed":   true,
	}
	resp = doRequest(t, "PUT", fmt.Sprintf("/tasks/%d", int(taskID)), updateBody, true)
	if resp.StatusCode != 200 {
		t.Errorf("PUT /tasks/%d: ожидался 200, получили %d", int(taskID), resp.StatusCode)
	}
	resp.Body.Close()

	// 5. Удаление задачи
	resp = doRequest(t, "DELETE", fmt.Sprintf("/tasks/%d", int(taskID)), nil, true)
	if resp.StatusCode != 204 {
		t.Errorf("DELETE /tasks/%d: ожидался 204, получили %d", int(taskID), resp.StatusCode)
	}
	resp.Body.Close()
}
