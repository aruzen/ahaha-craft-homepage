package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"backend/internal/domain"
	"backend/internal/service"
	"backend/pkg/api"
)

type DocsService interface {
	ListVaults(ctx context.Context) ([]domain.DocVault, error)
	ListNotes(ctx context.Context, vaultSlug, tag, group string) ([]domain.DocNote, error)
	GetNote(ctx context.Context, vaultSlug, noteSlug string) (domain.DocNote, error)
	ReadNoteContent(ctx context.Context, vaultSlug, noteSlug string) ([]byte, string, error)
	ReadAsset(ctx context.Context, vaultSlug, assetPath string) ([]byte, string, error)
	AdminListBranches(ctx context.Context, session domain.SessionData) ([]string, error)
	AdminListVaults(ctx context.Context, session domain.SessionData) ([]domain.DocVault, error)
	AdminRegisterVault(ctx context.Context, session domain.SessionData, branch, slug, title string) (domain.DocVault, error)
	AdminDisableVault(ctx context.Context, session domain.SessionData, vaultSlug string) error
	AdminSyncVault(ctx context.Context, session domain.SessionData, vaultSlug string) error
	AdminRescanVault(ctx context.Context, session domain.SessionData, vaultSlug string) error
	AdminOverrideNotePublished(ctx context.Context, session domain.SessionData, vaultSlug, noteSlug string, published bool) error
}

type DocsHandler struct {
	service DocsService
}

func NewDocsHandler(service DocsService) *DocsHandler {
	return &DocsHandler{service: service}
}

func (h *DocsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/docs/")
	switch {
	case r.URL.Path == "/api/docs/vaults":
		h.handleVaults(w, r)
	case strings.HasPrefix(path, "vaults/"):
		h.handleVaultScoped(w, r, strings.TrimPrefix(path, "vaults/"))
	case strings.HasPrefix(path, "content/"):
		h.handleContent(w, r, strings.TrimPrefix(path, "content/"))
	case strings.HasPrefix(path, "assets/"):
		h.handleAsset(w, r, strings.TrimPrefix(path, "assets/"))
	case strings.HasPrefix(path, "admin/"):
		h.handleAdmin(w, r, strings.TrimPrefix(path, "admin/"))
	default:
		respondNotFound(w)
	}
}

func (h *DocsHandler) handleVaults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondMethodNotAllowed(w, http.MethodGet)
		return
	}
	vaults, err := h.service.ListVaults(r.Context())
	if err != nil {
		handleDocsError(w, err)
		return
	}
	payloads := make([]api.DocVaultPayload, len(vaults))
	for i, vault := range vaults {
		payloads[i] = api.NewDocVaultPayload(vault, false)
	}
	respondJSON(w, http.StatusOK, api.DocVaultsResponse{Vaults: payloads})
}

func (h *DocsHandler) handleVaultScoped(w http.ResponseWriter, r *http.Request, rest string) {
	parts := strings.Split(rest, "/")
	if len(parts) < 2 || parts[1] != "notes" {
		respondNotFound(w)
		return
	}
	vaultSlug := parts[0]
	if len(parts) == 2 {
		if r.Method != http.MethodGet {
			respondMethodNotAllowed(w, http.MethodGet)
			return
		}
		notes, err := h.service.ListNotes(r.Context(), vaultSlug, r.URL.Query().Get("tag"), r.URL.Query().Get("group"))
		if err != nil {
			handleDocsError(w, err)
			return
		}
		payloads := make([]api.DocNotePayload, len(notes))
		for i, note := range notes {
			payloads[i] = api.NewDocNotePayload(note, false)
		}
		respondJSON(w, http.StatusOK, api.DocNotesResponse{Notes: payloads})
		return
	}
	if len(parts) == 3 {
		if r.Method != http.MethodGet {
			respondMethodNotAllowed(w, http.MethodGet)
			return
		}
		note, err := h.service.GetNote(r.Context(), vaultSlug, parts[2])
		if err != nil {
			handleDocsError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, api.NewDocNotePayload(note, false))
		return
	}
	respondNotFound(w)
}

func (h *DocsHandler) handleContent(w http.ResponseWriter, r *http.Request, rest string) {
	if r.Method != http.MethodGet {
		respondMethodNotAllowed(w, http.MethodGet)
		return
	}
	vaultSlug, noteSlug, ok := strings.Cut(rest, "/")
	if !ok {
		respondNotFound(w)
		return
	}
	body, contentType, err := h.service.ReadNoteContent(r.Context(), vaultSlug, noteSlug)
	if err != nil {
		handleDocsError(w, err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func (h *DocsHandler) handleAsset(w http.ResponseWriter, r *http.Request, rest string) {
	if r.Method != http.MethodGet {
		respondMethodNotAllowed(w, http.MethodGet)
		return
	}
	vaultSlug, assetPath, ok := strings.Cut(rest, "/")
	if !ok {
		respondNotFound(w)
		return
	}
	body, contentType, err := h.service.ReadAsset(r.Context(), vaultSlug, assetPath)
	if err != nil {
		handleDocsError(w, err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func (h *DocsHandler) handleAdmin(w http.ResponseWriter, r *http.Request, rest string) {
	if r.Method != http.MethodPost {
		respondMethodNotAllowed(w, http.MethodPost)
		return
	}
	switch {
	case rest == "branches":
		var req api.DocAdminSessionRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		session, err := req.Session.ToDomain()
		if err != nil {
			respondInvalidField(w, "session")
			return
		}
		branches, err := h.service.AdminListBranches(r.Context(), session)
		if err != nil {
			handleDocsError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, api.DocBranchesResponse{Branches: branches})
	case rest == "vaults":
		var req api.DocAdminSessionRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		session, err := req.Session.ToDomain()
		if err != nil {
			respondInvalidField(w, "session")
			return
		}
		vaults, err := h.service.AdminListVaults(r.Context(), session)
		if err != nil {
			handleDocsError(w, err)
			return
		}
		payloads := make([]api.DocVaultPayload, len(vaults))
		for i, vault := range vaults {
			payloads[i] = api.NewDocVaultPayload(vault, true)
		}
		respondJSON(w, http.StatusOK, api.DocVaultsResponse{Vaults: payloads})
	case rest == "vaults/register":
		var req api.DocRegisterVaultRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		session, err := req.Session.ToDomain()
		if err != nil {
			respondInvalidField(w, "session")
			return
		}
		vault, err := h.service.AdminRegisterVault(r.Context(), session, req.Branch, req.Slug, req.Title)
		if err != nil {
			handleDocsError(w, err)
			return
		}
		respondJSON(w, http.StatusCreated, api.NewDocVaultPayload(vault, true))
	case strings.HasPrefix(rest, "vaults/"):
		h.handleAdminVaultAction(w, r, strings.TrimPrefix(rest, "vaults/"))
	default:
		respondNotFound(w)
	}
}

func (h *DocsHandler) handleAdminVaultAction(w http.ResponseWriter, r *http.Request, rest string) {
	parts := strings.Split(rest, "/")
	if len(parts) < 2 {
		respondNotFound(w)
		return
	}
	vaultSlug := parts[0]
	if len(parts) == 2 {
		var req api.DocVaultActionRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		session, err := req.Session.ToDomain()
		if err != nil {
			respondInvalidField(w, "session")
			return
		}
		switch parts[1] {
		case "sync":
			err = h.service.AdminSyncVault(r.Context(), session, vaultSlug)
		case "rescan":
			err = h.service.AdminRescanVault(r.Context(), session, vaultSlug)
		case "disable":
			err = h.service.AdminDisableVault(r.Context(), session, vaultSlug)
		default:
			respondNotFound(w)
			return
		}
		if err != nil {
			handleDocsError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	if len(parts) == 4 && parts[1] == "notes" && parts[3] == "published" {
		var req api.DocOverridePublishedRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		session, err := req.Session.ToDomain()
		if err != nil {
			respondInvalidField(w, "session")
			return
		}
		if err := h.service.AdminOverrideNotePublished(r.Context(), session, vaultSlug, parts[2], req.Published); err != nil {
			handleDocsError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	respondNotFound(w)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dest any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		respondInvalidJSON(w)
		return false
	}
	return true
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondNotFound(w http.ResponseWriter) {
	respondAPIError(w, http.StatusNotFound, "not_found", "path", "resource not found")
}

func handleDocsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrDocInvalidInput):
		respondInvalidField(w, "docs")
	case errors.Is(err, service.ErrDocForbidden), errors.Is(err, domain.ErrInvalidLoginSession), errors.Is(err, domain.ErrExpiredToken):
		respondUnauthorizedSession(w)
	case errors.Is(err, service.ErrDocNotFound):
		respondNotFound(w)
	case errors.Is(err, service.ErrDocNotConfigured):
		respondAPIError(w, http.StatusServiceUnavailable, "not_configured", "docs", "DOC_VAULT_REPO_PATH is required")
	default:
		respondInternalServerError(w)
	}
}
