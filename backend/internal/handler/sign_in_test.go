package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/domain"
	"backend/pkg/api"

	"github.com/google/uuid"
)

func TestSignInHandler_ServeHTTP_Success(t *testing.T) {
	token, err := domain.NewLoginSessionToken()
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}
	session, err := domain.NewSessionData(uuid.New(), token)
	if err != nil {
		t.Fatalf("failed to create session data: %v", err)
	}

	svc := &fakeSignInService{session: session}
	handler := NewSignInHandler(svc)

	body := `{"name":"alice","email":"alice@example.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sign-in", strings.NewReader(body))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	if contentType := res.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected application/json, got %s", contentType)
	}

	var resp api.SignInResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Token != token.String() {
		t.Fatalf("expected token %s, got %s", token, resp.Token)
	}

	if resp.UserID != session.UserID().String() {
		t.Fatalf("expected user_id %s, got %s", session.UserID(), resp.UserID)
	}

	if !svc.called {
		t.Fatalf("service.SignIn was not called")
	}
}

func TestSignInHandler_InvalidJSON(t *testing.T) {
	svc := &fakeSignInService{}
	handler := NewSignInHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/sign-in", strings.NewReader(`{"name":1}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}

	if svc.called {
		t.Fatalf("service should not be called on invalid json")
	}
}

func TestSignInHandler_InvalidDomain(t *testing.T) {
	svc := &fakeSignInService{}
	handler := NewSignInHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/sign-in", strings.NewReader(`{"name":" ","email":"bad","password":"secret"}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}

	if svc.called {
		t.Fatalf("service should not be called on invalid domain input")
	}
}

func TestSignInHandler_Conflict(t *testing.T) {
	svc := &fakeSignInService{err: domain.ErrInvalidCredential}
	handler := NewSignInHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/sign-in", strings.NewReader(`{"name":"alice","email":"alice@example.com","password":"secret"}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", res.Code)
	}
}

func TestSignInHandler_InternalError(t *testing.T) {
	svc := &fakeSignInService{err: errors.New("boom")}
	handler := NewSignInHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/sign-in", strings.NewReader(`{"name":"alice","email":"alice@example.com","password":"secret"}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", res.Code)
	}
}

func TestSignInHandler_MethodNotAllowed(t *testing.T) {
	svc := &fakeSignInService{}
	handler := NewSignInHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/sign-in", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", res.Code)
	}

	if allow := res.Header().Get("Allow"); allow != http.MethodPost {
		t.Fatalf("expected Allow %s, got %s", http.MethodPost, allow)
	}
}

type fakeSignInService struct {
	session domain.SessionData
	role    domain.UserRole
	err     error
	called  bool
}

func (f *fakeSignInService) SignIn(_ context.Context, _ domain.SignInCredential) (domain.SessionData, domain.UserRole, error) {
	f.called = true
	if f.err != nil {
		return domain.SessionData{}, "", f.err
	}
	role := f.role
	if role == "" {
		role = domain.UserRoleAdmin
	}
	return f.session, role, nil
}
