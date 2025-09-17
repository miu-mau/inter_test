package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// -------------------- USERS --------------------

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	IsActive     bool   `json:"isActive"`
	IsAdmin      bool   `json:"isAdmin"`
}

type UsersStore struct {
	mu         sync.RWMutex
	usersByID  map[int64]*User
	byUsername map[string]*User
	byEmail    map[string]*User
	next       int64
}

func NewUsersStore() *UsersStore {
	return &UsersStore{
		usersByID:  map[int64]*User{},
		byUsername: map[string]*User{},
		byEmail:    map[string]*User{},
		next:       1,
	}
}

func (s *UsersStore) Create(u *User) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byUsername[strings.ToLower(u.Username)]; ok {
		return nil, errors.New("username already exists")
	}
	if _, ok := s.byEmail[strings.ToLower(u.Email)]; ok {
		return nil, errors.New("email already exists")
	}
	u.ID = s.next
	s.next++
	s.usersByID[u.ID] = u
	s.byUsername[strings.ToLower(u.Username)] = u
	s.byEmail[strings.ToLower(u.Email)] = u
	return u, nil
}

func (s *UsersStore) GetByUsername(username string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byUsername[strings.ToLower(username)]
	return u, ok
}

func (s *UsersStore) GetByID(id int64) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.usersByID[id]
	return u, ok
}

func (s *UsersStore) GetAll() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*User, 0, len(s.usersByID))
	for _, u := range s.usersByID {
		out = append(out, &User{ID: u.ID, Username: u.Username, Email: u.Email, IsActive: u.IsActive, IsAdmin: u.IsAdmin})
	}
	return out
}

func (s *UsersStore) Update(id int64, upd *User) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.usersByID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	// handle username uniqueness
	if upd.Username != "" && !strings.EqualFold(upd.Username, u.Username) {
		if _, exists := s.byUsername[strings.ToLower(upd.Username)]; exists {
			return nil, errors.New("username already exists")
		}
		delete(s.byUsername, strings.ToLower(u.Username))
		u.Username = upd.Username
		s.byUsername[strings.ToLower(u.Username)] = u
	}
	// handle email uniqueness
	if upd.Email != "" && !strings.EqualFold(upd.Email, u.Email) {
		if _, exists := s.byEmail[strings.ToLower(upd.Email)]; exists {
			return nil, errors.New("email already exists")
		}
		delete(s.byEmail, strings.ToLower(u.Email))
		u.Email = upd.Email
		s.byEmail[strings.ToLower(u.Email)] = u
	}
	// flags (caller already prepared desired values)
	u.IsActive = upd.IsActive
	u.IsAdmin = upd.IsAdmin
	// Note: password is not updated here
	return &User{ID: u.ID, Username: u.Username, Email: u.Email, IsActive: u.IsActive, IsAdmin: u.IsAdmin}, nil
}

func (s *UsersStore) Delete(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.usersByID[id]
	if !ok {
		return false
	}
	delete(s.byUsername, strings.ToLower(u.Username))
	delete(s.byEmail, strings.ToLower(u.Email))
	delete(s.usersByID, id)
	return true
}

// -------------------- TOKENS --------------------

type RefreshEntry struct {
	UserID    int64
	ExpiresAt time.Time
}

type RefreshStore struct {
	mu   sync.RWMutex
	data map[string]RefreshEntry // jti -> entry
}

func NewRefreshStore() *RefreshStore {
	return &RefreshStore{data: map[string]RefreshEntry{}}
}

func (s *RefreshStore) Add(jti string, userID int64, exp time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[jti] = RefreshEntry{UserID: userID, ExpiresAt: exp}
}

func (s *RefreshStore) Remove(jti string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, jti)
}

func (s *RefreshStore) IsValid(jti string) (RefreshEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.data[jti]
	if !ok {
		return RefreshEntry{}, false
	}
	if time.Now().After(e.ExpiresAt) {
		return RefreshEntry{}, false
	}
	return e, true
}

var jwtSecret = []byte("dev_super_secret_change_me")
var accessTTL = 15 * time.Minute
var refreshTTL = 7 * 24 * time.Hour

func generateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func issueToken(user *User, ttl time.Duration, tokenType string, jti string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(ttl)
	claims := jwt.MapClaims{
		"sub":      fmt.Sprintf("%d", user.ID),
		"username": user.Username,
		"type":     tokenType,
		"exp":      exp.Unix(),
		"iat":      now.Unix(),
	}
	if jti != "" {
		claims["jti"] = jti
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(jwtSecret)
	return s, exp, err
}

func parseAndValidateAccessToken(r *http.Request) (*jwt.Token, jwt.MapClaims, error) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, nil, errors.New("missing bearer token")
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, errors.New("invalid claims")
	}
	if t, ok := claims["type"].(string); !ok || t != "access" {
		return nil, nil, errors.New("invalid token type")
	}
	return token, claims, nil
}

// -------------------- HTTP --------------------

type RegisterPayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int64  `json:"expiresIn"`
	User         *User  `json:"user,omitempty"`
}

type refreshPayload struct {
	RefreshToken string `json:"refreshToken"`
}

func registerAuthRoutes(r *mux.Router, users *UsersStore, refreshes *RefreshStore) {
	api := r.PathPrefix("/api").Subrouter()
	auth := api.PathPrefix("/auth").Subrouter()

	// POST /api/auth/register
	auth.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		var p RegisterPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		p.Username = strings.TrimSpace(p.Username)
		p.Email = strings.TrimSpace(p.Email)
		if p.Username == "" || p.Email == "" || strings.TrimSpace(p.Password) == "" {
			http.Error(w, "username, email and password are required", http.StatusBadRequest)
			return
		}
		if len(p.Password) < 6 {
			http.Error(w, "password must be at least 6 characters", http.StatusBadRequest)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
			return
		}
		user := &User{
			Username:     p.Username,
			Email:        p.Email,
			PasswordHash: string(hash),
			IsActive:     true,
			IsAdmin:      false,
		}
		created, err := users.Create(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, &User{ID: created.ID, Username: created.Username, Email: created.Email, IsActive: created.IsActive, IsAdmin: created.IsAdmin})
	}).Methods(http.MethodPost, http.MethodOptions)

	// POST /api/auth/login
	auth.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var p LoginPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		u, ok := users.GetByUsername(strings.TrimSpace(p.Username))
		if !ok {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(p.Password)); err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		jti, err := generateRandomHex(16)
		if err != nil {
			http.Error(w, "failed to create token", http.StatusInternalServerError)
			return
		}
		access, _, err := issueToken(u, accessTTL, "access", "")
		if err != nil {
			http.Error(w, "failed to create token", http.StatusInternalServerError)
			return
		}
		refresh, rExp, err := issueToken(u, refreshTTL, "refresh", jti)
		if err != nil {
			http.Error(w, "failed to create token", http.StatusInternalServerError)
			return
		}
		refreshes.Add(jti, u.ID, rExp)
		writeJSON(w, &TokenResponse{
			AccessToken:  access,
			RefreshToken: refresh,
			TokenType:    "Bearer",
			ExpiresIn:    int64(accessTTL.Seconds()),
			User:         &User{ID: u.ID, Username: u.Username, Email: u.Email, IsActive: u.IsActive, IsAdmin: u.IsAdmin},
		})
	}).Methods(http.MethodPost, http.MethodOptions)

	// POST /api/auth/refresh
	auth.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		var p refreshPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil || strings.TrimSpace(p.RefreshToken) == "" {
			http.Error(w, "refreshToken is required", http.StatusBadRequest)
			return
		}
		token, err := jwt.Parse(p.RefreshToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
			return
		}
		if t, ok := claims["type"].(string); !ok || t != "refresh" {
			http.Error(w, "invalid token type", http.StatusUnauthorized)
			return
		}
		jti, _ := claims["jti"].(string)
		if jti == "" {
			http.Error(w, "invalid refresh token", http.StatusUnauthorized)
			return
		}
		entry, ok := refreshes.IsValid(jti)
		if !ok {
			http.Error(w, "refresh token expired or revoked", http.StatusUnauthorized)
			return
		}
		u, ok := users.GetByID(entry.UserID)
		if !ok {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}
		// rotate refresh token
		refreshes.Remove(jti)
		newJTI, err := generateRandomHex(16)
		if err != nil {
			http.Error(w, "failed to create token", http.StatusInternalServerError)
			return
		}
		access, _, err := issueToken(u, accessTTL, "access", "")
		if err != nil {
			http.Error(w, "failed to create token", http.StatusInternalServerError)
			return
		}
		refresh, rExp, err := issueToken(u, refreshTTL, "refresh", newJTI)
		if err != nil {
			http.Error(w, "failed to create token", http.StatusInternalServerError)
			return
		}
		refreshes.Add(newJTI, u.ID, rExp)
		writeJSON(w, &TokenResponse{
			AccessToken:  access,
			RefreshToken: refresh,
			TokenType:    "Bearer",
			ExpiresIn:    int64(accessTTL.Seconds()),
		})
	}).Methods(http.MethodPost, http.MethodOptions)

	// POST /api/auth/logout
	auth.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		var p refreshPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil || strings.TrimSpace(p.RefreshToken) == "" {
			http.Error(w, "refreshToken is required", http.StatusBadRequest)
			return
		}
		token, _ := jwt.Parse(p.RefreshToken, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if token != nil {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if jti, _ := claims["jti"].(string); jti != "" {
					refreshes.Remove(jti)
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}).Methods(http.MethodPost, http.MethodOptions)

	// GET /api/auth/me
	auth.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := parseAndValidateAccessToken(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		sub, _ := claims["sub"].(string)
		id, err := strconv.ParseInt(sub, 10, 64)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		u, ok := users.GetByID(id)
		if !ok {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}
		writeJSON(w, &User{ID: u.ID, Username: u.Username, Email: u.Email, IsActive: u.IsActive, IsAdmin: u.IsAdmin})
	}).Methods(http.MethodGet, http.MethodOptions)
}

// admin middleware util
func requireAdmin(users *UsersStore, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := parseAndValidateAccessToken(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		sub, _ := claims["sub"].(string)
		id, err := strconv.ParseInt(sub, 10, 64)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		u, ok := users.GetByID(id)
		if !ok || !u.IsAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// Admin endpoints under /api/users
func registerAdminRoutes(r *mux.Router, users *UsersStore) {
	api := r.PathPrefix("/api").Subrouter()
	usersR := api.PathPrefix("/users").Subrouter()

	// GET /api/users - list all users
	usersR.HandleFunc("", requireAdmin(users, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, users.GetAll())
	})).Methods(http.MethodGet, http.MethodOptions)

	// GET /api/users/{id}
	usersR.HandleFunc("/{id}", requireAdmin(users, func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		u, ok := users.GetByID(id)
		if !ok {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		writeJSON(w, &User{ID: u.ID, Username: u.Username, Email: u.Email, IsActive: u.IsActive, IsAdmin: u.IsAdmin})
	})).Methods(http.MethodGet, http.MethodOptions)

	// PUT /api/users/{id}
	type updateUserPayload struct {
		Username *string `json:"username,omitempty"`
		Email    *string `json:"email,omitempty"`
		IsActive *bool   `json:"isActive,omitempty"`
		IsAdmin  *bool   `json:"isAdmin,omitempty"`
	}
	usersR.HandleFunc("/{id}", requireAdmin(users, func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var p updateUserPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		u, ok := users.GetByID(id)
		if !ok {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		upd := &User{}
		if p.Username != nil {
			upd.Username = strings.TrimSpace(*p.Username)
		}
		if p.Email != nil {
			upd.Email = strings.TrimSpace(*p.Email)
		}
		if p.IsActive != nil {
			upd.IsActive = *p.IsActive
		} else {
			upd.IsActive = u.IsActive
		}
		if p.IsAdmin != nil {
			upd.IsAdmin = *p.IsAdmin
		} else {
			upd.IsAdmin = u.IsAdmin
		}
		updated, err := users.Update(id, upd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, updated)
	})).Methods(http.MethodPut, http.MethodOptions)

	// DELETE /api/users/{id}
	usersR.HandleFunc("/{id}", requireAdmin(users, func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		if !users.Delete(id) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})).Methods(http.MethodDelete, http.MethodOptions)
}

// seed default admin user (username: admin, password: admin123)
func seedDefaultAdmin(users *UsersStore) error {
	if _, ok := users.GetByUsername("admin"); ok {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = users.Create(&User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: string(hash),
		IsActive:     true,
		IsAdmin:      true,
	})
	return err
}
