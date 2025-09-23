package main

import "testing"

func TestStore_Create_Get_Update_Delete(t *testing.T) {
	s := NewStore()

	// create
	t1 := s.Create(&Task{Title: "A"})
	if t1.ID == 0 || t1.Title != "A" || t1.Completed {
		t.Fatalf("unexpected created: %+v", t1)
	}

	// get all
	all := s.GetAll()
	if len(all) != 1 || all[0].ID != t1.ID {
		t.Fatalf("unexpected all: %+v", all)
	}

	// get by id
	got, ok := s.GetByID(t1.ID)
	if !ok || got.Title != "A" {
		t.Fatalf("unexpected get: ok=%v got=%+v", ok, got)
	}

	// update title, description, completed
	newTitle := "B"
	desc := "desc"
	done := true
	upd := &UpdateTaskPayload{Title: &newTitle, Description: &desc, Completed: &done}
	updTask, err := s.Update(t1.ID, upd)
	if err != nil {
		t.Fatalf("update err: %v", err)
	}
	if updTask.Title != "B" || updTask.Description != "desc" || !updTask.Completed {
		t.Fatalf("unexpected updated: %+v", updTask)
	}

	// empty title should error
	empty := "  "
	_, err = s.Update(t1.ID, &UpdateTaskPayload{Title: &empty})
	if err == nil {
		t.Fatalf("expected error on empty title")
	}

	// delete
	if ok := s.Delete(t1.ID); !ok {
		t.Fatalf("delete returned false")
	}
	if _, ok := s.GetByID(t1.ID); ok {
		t.Fatalf("expected not found after delete")
	}
}
