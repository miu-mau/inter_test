package main

import (
	"net/http"
	"testing"
	"time"
)

func TestTokens_IssueAndValidateAccess(t *testing.T) {
	user := &User{ID: 7, Username: "u7"}
	// issue access token
	access, _, err := issueToken(user, 2*time.Second, "access", "")
	if err != nil || access == "" {
		t.Fatalf("issue access failed: %v", err)
	}

	// prepare request with bearer
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+access)

	_, claims, err := parseAndValidateAccessToken(r)
	if err != nil {
		t.Fatalf("validate access failed: %v", err)
	}
	if claims["username"].(string) != "u7" || claims["type"].(string) != "access" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}
