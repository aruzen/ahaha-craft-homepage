package api

import (
	"backend/internal/domain"
)

// SessionPayload は session-data-struct を JSON で表現する。
type SessionPayload struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

func NewSessionPayload(session domain.SessionData) SessionPayload {
	return SessionPayload{
		UserID: session.UserID().String(),
		Token:  session.Token().String(),
	}
}
