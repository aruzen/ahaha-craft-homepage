package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"backend/internal/domain"
	"backend/pkg/api"
)

// LoginService は認証処理を司るユースケース層の抽象インターフェース。
type LoginService interface {
	Login(ctx context.Context, credential domain.AdminCredential) (domain.SessionData, domain.UserRole, error)
}

// LoginHandler は /api/login の HTTP リクエストを処理する。
type LoginHandler struct {
	service LoginService
}

// NewLoginHandler はログイン用ハンドラを初期化する。
func NewLoginHandler(service LoginService) *LoginHandler {
	return &LoginHandler{service: service}
}

// ServeHTTP は JSON リクエストをデコードし、ドメインに変換してサービスへ委譲する。
func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondMethodNotAllowed(w, http.MethodPost)
		return
	}

	var req api.LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respondInvalidJSON(w)
		return
	}

	credential, err := req.ToDomain()
	if err != nil {
		respondInvalidField(w, "credential")
		return
	}

	session, role, err := h.service.Login(r.Context(), credential)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredential) {
			respondInvalidCredential(w, http.StatusUnauthorized)
		} else {
			respondInternalServerError(w)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(api.NewLoginResponse(session, role))
}
