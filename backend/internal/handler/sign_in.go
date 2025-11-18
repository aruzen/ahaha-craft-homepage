package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"backend/internal/domain"
	"backend/pkg/api"
)

// SignInService はサインイン処理を司るユースケース層の抽象インターフェース。
type SignInService interface {
	SignIn(ctx context.Context, credential domain.SignInCredential) (domain.SessionData, error)
}

// SignInHandler は /api/sign-in の HTTP リクエストを処理する。
type SignInHandler struct {
	service SignInService
}

func NewSignInHandler(service SignInService) *SignInHandler {
	return &SignInHandler{service: service}
}

func (h *SignInHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req api.SignInRequest
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

	session, err := h.service.SignIn(r.Context(), credential)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredential) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(api.NewSignInResponse(session))
}
