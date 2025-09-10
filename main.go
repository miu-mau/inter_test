package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

type Task struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Completed   bool   `json:"completed"`
}

type CreateTaskPayload struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Completed   *bool  `json:"completed,omitempty"`
}

type UpdateTaskPayload struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}

type Store struct {
	mu    sync.RWMutex
	tasks map[int64]*Task
	next  int64
}

func NewStore() *Store {
	return &Store{
		tasks: map[int64]*Task{},
		next:  1,
	}
}

func (s *Store) Create(t *Task) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.ID = s.next
	s.next++
	s.tasks[t.ID] = t
	return t
}

func (s *Store) GetAll() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		out = append(out, t)
	}
	return out
}

func (s *Store) GetByID(id int64) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *Store) Update(id int64, upd *UpdateTaskPayload) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, errors.New("not found")
	}
	if upd.Title != nil {
		if strings.TrimSpace(*upd.Title) == "" {
			return nil, errors.New("title cannot be empty")
		}
		t.Title = *upd.Title
	}
	if upd.Description != nil {
		t.Description = *upd.Description
	}
	if upd.Completed != nil {
		t.Completed = *upd.Completed
	}
	return t, nil
}

func (s *Store) Delete(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return false
	}
	delete(s.tasks, id)
	return true
}

func main() {
	store := NewStore()
	r := mux.NewRouter()

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

	addr := ":8080"
	fmt.Printf("Server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
