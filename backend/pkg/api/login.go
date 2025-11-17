package api

import (
	"backend/internal/domain"
	"github.com/google/uuid"
)

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (r LoginRequest) ToDomain() (domain.AdminCredential, error) {
	name, err := domain.NewName(r.Name)
	if err != nil {
		return domain.AdminCredential{}, err
	}

	return domain.NewAdminCredential(name, r.Password)
}

type LoginResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

func NewLoginResponse(userID uuid.UUID, token domain.LoginSessionToken) LoginResponse {
	return LoginResponse{UserID: userID.String(), Token: token.String()}
}
