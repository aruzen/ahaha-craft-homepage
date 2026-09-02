package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/domain"
	"backend/internal/service"
	"backend/pkg/api"
)

type DocsService interface {
	ListVaults(ctx context.Context) ([]domain.DocVault, error)
	ListToys(ctx context.Context) ([]domain.DocToy, error)
	ListNotes(ctx context.Context, vaultSlug, tag, group string) ([]domain.DocNote, error)
	GetNote(ctx context.Context, vaultSlug, noteSlug string) (domain.DocNote, error)
	ReadNoteContent(ctx context.Context, vaultSlug, noteSlug string) ([]byte, string, error)
	ReadAsset(ctx context.Context, vaultSlug, assetPath string) ([]byte, string, error)
	AdminListBranches(ctx context.Context, session domain.SessionData) ([]string, error)
	AdminListVaults(ctx context.Context, session domain.SessionData) ([]domain.DocVault, error)
	AdminListNotes(ctx context.Context, session domain.SessionData, vaultSlug string) ([]domain.DocNote, error)
	AdminUpdateNoteMetadata(ctx context.Context, session domain.SessionData, vaultSlug, noteSlug string, override domain.DocNoteOverride) error
	AdminUpload(ctx context.Context, session domain.SessionData, upload service.DocUpload) error
	AdminTrashLocal(ctx context.Context, session domain.SessionData, noteSlug string) error
	AdminListTrash(ctx context.Context, session domain.SessionData) ([]string, error)
	AdminRestoreLocal(ctx context.Context, session domain.SessionData, trashPath string) error
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
	case r.URL.Path == "/api/docs/toys":
		h.handleToys(w, r)
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

func (h *DocsHandler) handleToys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondMethodNotAllowed(w, http.MethodGet)
		return
	}
	toys, err := h.service.ListToys(r.Context())
	if err != nil {
		handleDocsError(w, err)
		return
	}
	payloads := make([]api.DocToyPayload, len(toys))
	for i, toy := range toys {
		payloads[i] = api.DocToyPayload{Source: api.NewDocVaultPayload(toy.Vault, false), Note: api.NewDocNotePayload(toy.Note, false)}
	}
	respondJSON(w, http.StatusOK, api.DocToysResponse{Toys: payloads})
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
	case rest == "uploads":
		h.handleAdminUpload(w, r)
	case rest == "uploads/trash":
		var req struct {
			Session  api.DocSessionPayload `json:"session"`
			NoteSlug string                `json:"note_slug"`
		}
		if !decodeJSON(w, r, &req) {
			return
		}
		session, err := req.Session.ToDomain()
		if err != nil {
			respondInvalidField(w, "session")
			return
		}
		if err := h.service.AdminTrashLocal(r.Context(), session, req.NoteSlug); err != nil {
			handleDocsError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case rest == "uploads/trash/list":
		var req api.DocAdminSessionRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		session, err := req.Session.ToDomain()
		if err != nil {
			respondInvalidField(w, "session")
			return
		}
		items, err := h.service.AdminListTrash(r.Context(), session)
		if err != nil {
			handleDocsError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]any{"items": items})
	case rest == "uploads/restore":
		var req struct {
			Session api.DocSessionPayload `json:"session"`
			Path    string                `json:"path"`
		}
		if !decodeJSON(w, r, &req) {
			return
		}
		session, err := req.Session.ToDomain()
		if err != nil {
			respondInvalidField(w, "session")
			return
		}
		if err := h.service.AdminRestoreLocal(r.Context(), session, req.Path); err != nil {
			handleDocsError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
	if len(parts) == 2 && parts[1] == "notes" {
		var req api.DocAdminSessionRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		session, err := req.Session.ToDomain()
		if err != nil {
			respondInvalidField(w, "session")
			return
		}
		notes, err := h.service.AdminListNotes(r.Context(), session, vaultSlug)
		if err != nil {
			handleDocsError(w, err)
			return
		}
		payloads := make([]api.DocNotePayload, len(notes))
		for i, note := range notes {
			payloads[i] = api.NewDocNotePayload(note, true)
		}
		respondJSON(w, http.StatusOK, api.DocNotesResponse{Notes: payloads})
		return
	}
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
	if len(parts) == 4 && parts[1] == "notes" && parts[3] == "metadata" {
		var req api.DocUpdateMetadataRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		session, err := req.Session.ToDomain()
		if err != nil {
			respondInvalidField(w, "session")
			return
		}
		override := domain.DocNoteOverride{Title: req.Title, Summary: req.Summary, Published: req.Published, Order: req.Order, Group: req.Group}
		if req.Tags != nil {
			tags := make([]domain.DocTag, 0, len(*req.Tags))
			for _, name := range *req.Tags {
				name = strings.TrimSpace(name)
				slug := service.SafeDocSlug(name)
				if name != "" && slug != "" {
					tags = append(tags, domain.DocTag{Slug: slug, Name: name})
				}
			}
			override.Tags = &tags
		}
		if err := h.service.AdminUpdateNoteMetadata(r.Context(), session, vaultSlug, parts[2], override); err != nil {
			handleDocsError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	respondNotFound(w)
}

func (h *DocsHandler) handleAdminUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 52<<20)
	if err := r.ParseMultipartForm(52 << 20); err != nil {
		respondInvalidField(w, "file")
		return
	}
	var sessionPayload api.DocSessionPayload
	if err := json.Unmarshal([]byte(r.FormValue("session")), &sessionPayload); err != nil {
		respondInvalidField(w, "session")
		return
	}
	session, err := sessionPayload.ToDomain()
	if err != nil {
		respondInvalidField(w, "session")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondInvalidField(w, "file")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, (50<<20)+1))
	if err != nil || len(data) > 50<<20 {
		respondInvalidField(w, "file")
		return
	}
	overwrite, _ := strconv.ParseBool(r.FormValue("overwrite"))
	err = h.service.AdminUpload(r.Context(), session, service.DocUpload{Kind: r.FormValue("kind"), Slug: r.FormValue("slug"), Filename: header.Filename, Data: data, Overwrite: overwrite})
	if err != nil {
		handleDocsError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
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
	case errors.Is(err, service.ErrDocConflict):
		respondAPIError(w, http.StatusConflict, "content_exists", "file", "content already exists")
	default:
		respondInternalServerError(w)
	}
}
