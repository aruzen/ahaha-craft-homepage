package api

import "backend/internal/domain"

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
	UserName string           `json:"user_name"`
	Record   HueRecordPayload `json:"record"`
}

func (r SaveResultRequest) ToDomain() (domain.HueResultSubmission, error) {
	record, err := r.Record.ToDomain()
	if err != nil {
		return domain.HueResultSubmission{}, err
	}

	userData, err := domain.NewName(r.UserName)
	if err != nil {
		return domain.HueResultSubmission{}, err
	}

	return domain.NewHueResultSubmission(userData, record), nil
}

// SaveResultResponse は仕様上ボディ不要のため空。
type SaveResultResponse struct{}

type GetDataRequest struct {
	Token     string `json:"token"`
	DataRange []int  `json:"data-range"`
}

func (r GetDataRequest) ToDomain() (domain.LoginSessionToken, domain.RecordRange, error) {
	token, err := domain.NewLoginSessionTokenFromPersistence(r.Token)
	if err != nil {
		return domain.LoginSessionToken{}, domain.RecordRange{}, err
	}

	if len(r.DataRange) != 2 {
		return domain.LoginSessionToken{}, domain.RecordRange{}, domain.ErrInvalidRange
	}

	recordRange, err := domain.NewRecordRange(r.DataRange[0], r.DataRange[1])
	if err != nil {
		return domain.LoginSessionToken{}, domain.RecordRange{}, err
	}

	return token, recordRange, nil
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
