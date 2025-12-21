package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"backend/internal/domain"
	"backend/pkg/api"
)

// HueSaveService は Hue 結果保存のユースケース境界。
type HueSaveService interface {
	SaveResult(ctx context.Context, record domain.HueRecord) (domain.HueResult, error)
}

// HueGetService は Hue データ取得のユースケース境界。
type HueGetService interface {
	GetData(ctx context.Context, session domain.SessionData, recordRange domain.RecordRange) ([]domain.HueRecord, error)
}

type HueSaveHandler struct {
	service HueSaveService
}

func NewHueSaveHandler(service HueSaveService) *HueSaveHandler {
	return &HueSaveHandler{service: service}
}

func (h *HueSaveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondMethodNotAllowed(w, http.MethodPost)
		return
	}

	var req api.SaveResultRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		log.Print("error: ", err)
		respondInvalidJSON(w)
		return
	}

	submission, err := req.ToDomain()
	if err != nil {
		log.Print("error: ", err)
		respondInvalidField(w, "record")
		return
	}

	result, err := h.service.SaveResult(r.Context(), submission)
	if err != nil {
		handleHueServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(api.NewSaveResultResponse(result))
}

type HueGetHandler struct {
	service HueGetService
}

func NewHueGetHandler(service HueGetService) *HueGetHandler {
	return &HueGetHandler{service: service}
}

func (h *HueGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondMethodNotAllowed(w, http.MethodPost)
		return
	}

	var req api.GetDataRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respondInvalidJSON(w)
		return
	}

	session, recordRange, err := req.ToDomain()
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidSessionToken),
			errors.Is(err, domain.ErrInvalidSessionData):
			respondInvalidField(w, "session")
		case errors.Is(err, domain.ErrInvalidRange):
			respondInvalidField(w, "data-range")
		default:
			respondInvalidField(w, "request")
		}
		return
	}

	records, err := h.service.GetData(r.Context(), session, recordRange)
	if err != nil {
		handleHueServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(api.NewGetDataResponse(records))
}

func handleHueServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidSessionToken),
		errors.Is(err, domain.ErrInvalidLoginSession),
		errors.Is(err, domain.ErrExpiredToken):
		respondUnauthorizedSession(w)
	default:
		respondInternalServerError(w)
	}
}
