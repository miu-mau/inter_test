package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
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
	r := BuildRouter()
	addr := ":8080"
	fmt.Printf("Server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
