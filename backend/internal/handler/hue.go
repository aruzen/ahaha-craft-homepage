package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"backend/internal/domain"
	"backend/pkg/api"
)

// HueSaveService は Hue 結果保存のユースケース境界。
type HueSaveService interface {
	SaveResult(ctx context.Context, record domain.HueRecord) error
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
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req api.SaveResultRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	submission, err := req.ToDomain()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := h.service.SaveResult(r.Context(), submission); err != nil {
		handleHueServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type HueGetHandler struct {
	service HueGetService
}

func NewHueGetHandler(service HueGetService) *HueGetHandler {
	return &HueGetHandler{service: service}
}

func (h *HueGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req api.GetDataRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	session, recordRange, err := req.ToDomain()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
