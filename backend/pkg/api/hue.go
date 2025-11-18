package api

import (
	"backend/internal/domain"
	"github.com/google/uuid"
)

// HueRecordPayload は hue-are-you の回答を JSON で表す。
type HueRecordPayload struct {
	Name   string            `json:"name"`
	Choice map[string]string `json:"choice"`
}

func (p HueRecordPayload) ToDomain() (domain.HueRecord, error) {
	return domain.NewHueRecordFromRaw(p.Name, p.Choice)
}

func NewHueRecordPayload(record domain.HueRecord) HueRecordPayload {
	return HueRecordPayload{
		Name:   record.Name().String(),
		Choice: record.ChoiceMap(),
	}
}

type SaveResultRequest struct {
	Session SessionPayload   `json:"session"`
	Record  HueRecordPayload `json:"record"`
}

func (r SaveResultRequest) ToDomain() (domain.HueResultSubmission, error) {
	record, err := r.Record.ToDomain()
	if err != nil {
		return domain.HueResultSubmission{}, err
	}

	userID, err := uuid.Parse(r.Session.UserID)
	if err != nil {
		return domain.HueResultSubmission{}, err
	}
	token, err := domain.ParseLoginSessionToken(r.Session.Token)
	if err != nil {
		return domain.HueResultSubmission{}, err
	}
	sessionData, err := domain.NewSessionData(userID, token)
	if err != nil {
		return domain.HueResultSubmission{}, err
	}

	return domain.NewHueResultSubmission(sessionData, record), nil
}

// SaveResultResponse は仕様上ボディ不要のため空。
type SaveResultResponse struct{}

type GetDataRequest struct {
	Session   SessionPayload `json:"session"`
	DataRange []int          `json:"data-range"`
}

func (r GetDataRequest) ToDomain() (domain.SessionData, domain.RecordRange, error) {
	id, err := uuid.Parse(r.Session.UserID)
	if err != nil {
		return domain.SessionData{}, domain.RecordRange{}, err
	}

	token, err := domain.ParseLoginSessionToken(r.Session.Token)
	if err != nil {
		return domain.SessionData{}, domain.RecordRange{}, err
	}

	if len(r.DataRange) != 2 {
		return domain.SessionData{}, domain.RecordRange{}, domain.ErrInvalidRange
	}

	recordRange, err := domain.NewRecordRange(r.DataRange[0], r.DataRange[1])
	if err != nil {
		return domain.SessionData{}, domain.RecordRange{}, err
	}

	session, err := domain.NewSessionData(id, token)
	if err != nil {
		return domain.SessionData{}, domain.RecordRange{}, err
	}

	return session, recordRange, nil
}

type GetDataResponse struct {
	Records []HueRecordPayload `json:"records"`
}

func NewGetDataResponse(records []domain.HueRecord) GetDataResponse {
	payloads := make([]HueRecordPayload, len(records))
	for i, record := range records {
		payloads[i] = NewHueRecordPayload(record)
	}

	return GetDataResponse{Records: payloads}
}
