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

func TestHueSaveHandler_ServeHTTP_Success(t *testing.T) {
	token, _ := domain.NewLoginSessionToken()
	session, _ := domain.NewSessionData(uuid.New(), token)
	record := buildHueRecord(t)
	submission := domain.NewHueResultSubmission(session, record)

	svc := &fakeHueSaveService{submission: submission}
	handler := NewHueSaveHandler(svc)

	reqBody := marshal(t, api.SaveResultRequest{
		Session: api.NewSessionPayload(session),
		Record:  api.NewHueRecordPayload(record),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/hue/save", strings.NewReader(reqBody))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.Code)
	}

	if !svc.called {
		t.Fatalf("service.SaveResult not called")
	}
}

func TestHueSaveHandler_InvalidJSON(t *testing.T) {
	handler := NewHueSaveHandler(&fakeHueSaveService{})
	req := httptest.NewRequest(http.MethodPost, "/api/hue/save", strings.NewReader(`{"session":1}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestHueSaveHandler_InvalidDomain(t *testing.T) {
	handler := NewHueSaveHandler(&fakeHueSaveService{})
	req := httptest.NewRequest(http.MethodPost, "/api/hue/save", strings.NewReader(`{"session":{"user_id":"bad","token":""},"record":{"name":"a","choice":{"w":"赤"}}}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestHueSaveHandler_Unauthorized(t *testing.T) {
	svc := &fakeHueSaveService{err: domain.ErrInvalidSessionToken}
	handler := NewHueSaveHandler(svc)

	token, _ := domain.NewLoginSessionToken()
	session, _ := domain.NewSessionData(uuid.New(), token)
	record := buildHueRecord(t)
	req := httptest.NewRequest(http.MethodPost, "/api/hue/save", strings.NewReader(marshal(t, api.SaveResultRequest{
		Session: api.NewSessionPayload(session),
		Record:  api.NewHueRecordPayload(record),
	})))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}

func TestHueSaveHandler_InternalError(t *testing.T) {
	svc := &fakeHueSaveService{err: errors.New("boom")}
	handler := NewHueSaveHandler(svc)

	token, _ := domain.NewLoginSessionToken()
	session, _ := domain.NewSessionData(uuid.New(), token)
	record := buildHueRecord(t)
	req := httptest.NewRequest(http.MethodPost, "/api/hue/save", strings.NewReader(marshal(t, api.SaveResultRequest{
		Session: api.NewSessionPayload(session),
		Record:  api.NewHueRecordPayload(record),
	})))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", res.Code)
	}
}

func TestHueSaveHandler_MethodNotAllowed(t *testing.T) {
	handler := NewHueSaveHandler(&fakeHueSaveService{})
	req := httptest.NewRequest(http.MethodGet, "/api/hue/save", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", res.Code)
	}
}

func TestHueGetHandler_ServeHTTP_Success(t *testing.T) {
	token, _ := domain.NewLoginSessionToken()
	session, _ := domain.NewSessionData(uuid.New(), token)
	record := buildHueRecord(t)
	svc := &fakeHueGetService{records: []domain.HueRecord{record}}
	handler := NewHueGetHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/hue/get", strings.NewReader(marshal(t, api.GetDataRequest{
		Session:   api.NewSessionPayload(session),
		DataRange: []int{0, 0},
	})))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	if contentType := res.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected application/json, got %s", contentType)
	}

	var resp api.GetDataResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Records) != 1 || resp.Records[0].Name != record.Name().String() {
		t.Fatalf("unexpected response payload")
	}
}

func TestHueGetHandler_InvalidJSON(t *testing.T) {
	handler := NewHueGetHandler(&fakeHueGetService{})
	req := httptest.NewRequest(http.MethodGet, "/api/hue/get", strings.NewReader(`{"session":1}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestHueGetHandler_InvalidDomain(t *testing.T) {
	handler := NewHueGetHandler(&fakeHueGetService{})
	req := httptest.NewRequest(http.MethodGet, "/api/hue/get", strings.NewReader(`{"session":{"user_id":"bad","token":""},"data-range":[0,0]}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestHueGetHandler_Unauthorized(t *testing.T) {
	svc := &fakeHueGetService{err: domain.ErrExpiredToken}
	handler := NewHueGetHandler(svc)
	token, _ := domain.NewLoginSessionToken()
	session, _ := domain.NewSessionData(uuid.New(), token)
	req := httptest.NewRequest(http.MethodGet, "/api/hue/get", strings.NewReader(marshal(t, api.GetDataRequest{
		Session:   api.NewSessionPayload(session),
		DataRange: []int{0, 0},
	})))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.Code)
	}
}

func TestHueGetHandler_InternalError(t *testing.T) {
	svc := &fakeHueGetService{err: errors.New("boom")}
	handler := NewHueGetHandler(svc)
	token, _ := domain.NewLoginSessionToken()
	session, _ := domain.NewSessionData(uuid.New(), token)
	req := httptest.NewRequest(http.MethodGet, "/api/hue/get", strings.NewReader(marshal(t, api.GetDataRequest{
		Session:   api.NewSessionPayload(session),
		DataRange: []int{0, 0},
	})))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", res.Code)
	}
}

func TestHueGetHandler_MethodNotAllowed(t *testing.T) {
	handler := NewHueGetHandler(&fakeHueGetService{})
	req := httptest.NewRequest(http.MethodPost, "/api/hue/get", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", res.Code)
	}
}

type fakeHueSaveService struct {
	submission domain.HueResultSubmission
	err        error
	called     bool
}

func (f *fakeHueSaveService) SaveResult(_ context.Context, submission domain.HueResultSubmission) error {
	f.called = true
	f.submission = submission
	if f.err != nil {
		return f.err
	}
	return nil
}

type fakeHueGetService struct {
	records []domain.HueRecord
	err     error
}

func (f *fakeHueGetService) GetData(_ context.Context, _ domain.SessionData, _ domain.RecordRange) ([]domain.HueRecord, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.records, nil
}

func marshal(t *testing.T, v interface{}) string {
	t.Helper()
	bytes, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	return string(bytes)
}

func buildHueRecord(t *testing.T) domain.HueRecord {
	t.Helper()
	name, err := domain.NewName("Tester")
	if err != nil {
		t.Fatalf("name error: %v", err)
	}
	choices, err := domain.NewHueChoices(map[string]string{"word": "赤"})
	if err != nil {
		t.Fatalf("choices error: %v", err)
	}
	record, err := domain.NewHueRecord(name, choices)
	if err != nil {
		t.Fatalf("record error: %v", err)
	}
	return record
}
