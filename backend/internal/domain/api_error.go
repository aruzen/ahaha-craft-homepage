package domain

import "strings"

// APIError は API エラー応答の値を保持する。
type APIError struct {
	cause   string
	field   string
	message string
}

// NewAPIError は cause/field/message をトリムし、いずれかが空なら ErrInvalidAPIError を返す。
func NewAPIError(cause, field, message string) (APIError, error) {
	c := strings.TrimSpace(cause)
	f := strings.TrimSpace(field)
	m := strings.TrimSpace(message)
	if c == "" || f == "" || m == "" {
		return APIError{}, ErrInvalidAPIError
	}

	return APIError{
		cause:   c,
		field:   f,
		message: m,
	}, nil
}

// Cause はエラー原因の識別子を返す。
func (e APIError) Cause() string {
	return e.cause
}

// Field は問題の対象フィールドを返す。
func (e APIError) Field() string {
	return e.field
}

// Message は表示用メッセージを返す。
func (e APIError) Message() string {
	return e.message
}
