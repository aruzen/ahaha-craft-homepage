package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/domain"
	"backend/pkg/api"
)

func TestLoginHandler_ServeHTTP_Success(t *testing.T) {
	token, err := domain.NewLoginSessionToken()
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	svc := &fakeLoginService{token: token}
	handler := NewLoginHandler(svc)

	name := "admin"
	password := "secret"
	body := fmt.Sprintf(`{"name":"%v","password":"%v"}`, name, password)

	req := httptest.NewRequest(http.MethodGet, "/api/login", strings.NewReader(body))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	if contentType := res.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", contentType)
	}

	var resp api.LoginResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Token != token.String() {
		t.Fatalf("expected token %s, got %s", token.String(), resp.Token)
	}

	if !svc.called {
		t.Fatalf("service.Login was not called")
	}

	if got := bcrypt.CompareHashAndPassword([]byte(svc.credential.HashedPassword()), []byte(password)); got != nil {
		t.Fatalf("expected hashed password 'secret', got %s", got)
	}

	if got := svc.credential.Name().String(); got != "admin" {
		t.Fatalf("expected name 'admin', got %s", got)
	}
}

func TestLoginHandler_ServeHTTP_InvalidJSON(t *testing.T) {
	svc := &fakeLoginService{}
	handler := NewLoginHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/login", strings.NewReader(`{"name":1}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.Code)
	}

	if svc.called {
		t.Fatalf("service should not be called on invalid JSON")
	}
}

func TestLoginHandler_ServeHTTP_InvalidDomainInput(t *testing.T) {
	svc := &fakeLoginService{}
	handler := NewLoginHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/login", strings.NewReader(`{"name":" ","hashed_password":"secret"}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.Code)
	}

	if svc.called {
		t.Fatalf("service should not be called on invalid domain input")
	}
}

func TestLoginHandler_ServeHTTP_InvalidCredential(t *testing.T) {
	svc := &fakeLoginService{err: domain.ErrInvalidCredential}
	handler := NewLoginHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/login", strings.NewReader(`{"name":"admin","password":"secret"}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}

func TestLoginHandler_ServeHTTP_InternalError(t *testing.T) {
	svc := &fakeLoginService{err: errors.New("boom")}
	handler := NewLoginHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/login", strings.NewReader(`{"name":"admin","password":"secret"}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", res.Code)
	}
}

func TestLoginHandler_ServeHTTP_MethodNotAllowed(t *testing.T) {
	svc := &fakeLoginService{}
	handler := NewLoginHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", res.Code)
	}

	if allow := res.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("expected Allow header %s, got %s", http.MethodGet, allow)
	}
}

type fakeLoginService struct {
	credential domain.AdminCredential
	userID     uuid.UUID
	token      domain.LoginSessionToken
	err        error
	called     bool
}

func (f *fakeLoginService) Login(_ context.Context, credential domain.AdminCredential) (uuid.UUID, domain.LoginSessionToken, error) {
	f.called = true
	f.credential = credential
	if f.err != nil {
		return uuid.Nil, domain.LoginSessionToken{}, f.err
	}
	return f.userID, f.token, nil
}
