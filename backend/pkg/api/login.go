package api

import "backend/internal/domain"

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
	SessionPayload
}

func NewLoginResponse(session domain.SessionData) LoginResponse {
	return LoginResponse{SessionPayload: NewSessionPayload(session)}
}
