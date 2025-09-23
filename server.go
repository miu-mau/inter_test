package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// BuildRouter constructs the HTTP router with all routes and in-memory stores.
// This enables top-down integration testing using httptest against the full stack.
func BuildRouter() *mux.Router {
	store := NewStore()
	r := mux.NewRouter()

	// shared auth stores
	users := NewUsersStore()
	refreshes := NewRefreshStore()
	_ = seedDefaultAdmin(users)

	r.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		// POST /tasks - создание
		if r.Method == http.MethodPost {
			var p CreateTaskPayload
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			p.Title = strings.TrimSpace(p.Title)
			if p.Title == "" {
				http.Error(w, "title is required", http.StatusBadRequest)
				return
			}
			task := &Task{
				Title:       p.Title,
				Description: p.Description,
				Completed:   false,
			}
			if p.Completed != nil {
				task.Completed = *p.Completed
			}
			store.Create(task)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(task)
			return
		}

		// GET /tasks - вывод списка (с возможностью фильтрации по completed)
		if r.Method == http.MethodGet {
			q := r.URL.Query().Get("completed")
			all := store.GetAll()
			if q == "" {
				writeJSON(w, all)
				return
			}
			// parse bool
			completed, err := strconv.ParseBool(q)
			if err != nil {
				http.Error(w, "completed query param must be boolean", http.StatusBadRequest)
				return
			}
			filtered := make([]*Task, 0)
			for _, t := range all {
				if t.Completed == completed {
					filtered = append(filtered, t)
				}
			}
			writeJSON(w, filtered)
			return
		}

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}).Methods(http.MethodPost, http.MethodGet, http.MethodOptions)

	// изменение по id
	r.HandleFunc("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			t, ok := store.GetByID(id)
			if !ok {
				http.Error(w, "task not found", http.StatusNotFound)
				return
			}
			writeJSON(w, t)
			return
		case http.MethodPut:
			var p UpdateTaskPayload
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			updated, err := store.Update(id, &p)
			if err != nil {
				if err.Error() == "not found" {
					http.Error(w, "task not found", http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, updated)
			return
		case http.MethodDelete:
			ok := store.Delete(id)
			if !ok {
				http.Error(w, "task not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}).Methods(http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodOptions)

	// auth and admin routes
	registerAuthRoutes(r, users, refreshes)
	registerAdminRoutes(r, users)

	return r
}


