package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net/http"

	"backend/internal/domain"
	"backend/pkg/api"
)

// LoginService は認証処理を司るユースケース層の抽象インターフェース。
type LoginService interface {
	Login(ctx context.Context, credential domain.AdminCredential) (uuid.UUID, domain.LoginSessionToken, error)
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
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req api.LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	credential, err := req.ToDomain()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, token, err := h.service.Login(r.Context(), credential)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredential) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(api.NewLoginResponse(id, token))
}
