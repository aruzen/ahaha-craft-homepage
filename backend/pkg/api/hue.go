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
	HueRecordPayload
}

func (r SaveResultRequest) ToDomain() (domain.HueRecord, error) {
	record, err := r.HueRecordPayload.ToDomain()
	if err != nil {
		return domain.HueRecord{}, err
	}
	return record, nil
}

type HuePayload struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

func NewHuePayload(h domain.HueRGB) HuePayload {
	return HuePayload{R: h.R(), G: h.G(), B: h.B()}
}

// SaveResultResponse は色とメッセージを返す。
type SaveResultResponse struct {
	Hue     HuePayload `json:"hue"`
	Message string     `json:"message"`
}

func NewSaveResultResponse(result domain.HueResult) SaveResultResponse {
	return SaveResultResponse{
		Hue:     NewHuePayload(result.Hue()),
		Message: result.Message(),
	}
}

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
