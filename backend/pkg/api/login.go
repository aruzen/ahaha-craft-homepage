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
	Role string `json:"role"`
}

func NewLoginResponse(session domain.SessionData, role domain.UserRole) LoginResponse {
	return LoginResponse{SessionPayload: NewSessionPayload(session), Role: role.String()}
}
