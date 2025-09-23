package main

import (
	"testing"
	"time"
)

func TestRefreshStore_Add_IsValid_Remove(t *testing.T) {
	rs := NewRefreshStore()

	jti := "abc"
	rs.Add(jti, 42, time.Now().Add(100*time.Millisecond))

	if _, ok := rs.IsValid(jti); !ok {
		t.Fatalf("expected valid right after add")
	}

	// after expiry should be invalid
	time.Sleep(120 * time.Millisecond)
	if _, ok := rs.IsValid(jti); ok {
		t.Fatalf("expected expired invalid")
	}

	// removing non-existent should be no panic
	rs.Remove(jti)
}
