package main

import "testing"

func TestUsersStore_CRUD_Uniqueness(t *testing.T) {
	s := NewUsersStore()

	// create user
	u, err := s.Create(&User{Username: "john", Email: "john@example.com", PasswordHash: "h", IsActive: true})
	if err != nil || u.ID == 0 {
		t.Fatalf("create failed: u=%+v err=%v", u, err)
	}

	// uniqueness username
	if _, err := s.Create(&User{Username: "john", Email: "john2@example.com", PasswordHash: "h"}); err == nil {
		t.Fatalf("expected username unique error")
	}
	// uniqueness email
	if _, err := s.Create(&User{Username: "john2", Email: "john@example.com", PasswordHash: "h"}); err == nil {
		t.Fatalf("expected email unique error")
	}

	// get by username
	got, ok := s.GetByUsername("john")
	if !ok || got.ID != u.ID {
		t.Fatalf("get by username failed: ok=%v got=%+v", ok, got)
	}

	// update: change username and email, keep flags
	upd, err := s.Update(u.ID, &User{Username: "johnny", Email: "johnny@example.com", IsActive: u.IsActive, IsAdmin: u.IsAdmin})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if upd.Username != "johnny" || upd.Email != "johnny@example.com" {
		t.Fatalf("unexpected updated: %+v", upd)
	}

	// delete
	if ok := s.Delete(u.ID); !ok {
		t.Fatalf("delete failed")
	}
	if _, ok := s.GetByID(u.ID); ok {
		t.Fatalf("expected user removed")
	}
}
