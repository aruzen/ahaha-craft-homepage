package api

import "backend/internal/domain"

// ErrorResponse は API のエラー応答を JSON 化したもの。
type ErrorResponse struct {
	Error   string `json:"error"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewErrorResponse はドメインの APIError から変換する。
func NewErrorResponse(err domain.APIError) ErrorResponse {
	return ErrorResponse{
		Error:   err.Cause(),
		Field:   err.Field(),
		Message: err.Message(),
	}
}

// ToDomain は JSON からドメイン型へ再変換する。
func (r ErrorResponse) ToDomain() (domain.APIError, error) {
	return domain.NewAPIError(r.Error, r.Field, r.Message)
}
