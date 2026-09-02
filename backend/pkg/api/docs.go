package api

import (
	"backend/internal/domain"

	"github.com/google/uuid"
)

type DocVaultPayload struct {
	Slug         string  `json:"slug"`
	Title        string  `json:"title"`
	Branch       string  `json:"branch,omitempty"`
	LocalPath    string  `json:"local_path,omitempty"`
	Status       string  `json:"status,omitempty"`
	LastSyncedAt *string `json:"last_synced_at,omitempty"`
}

type DocTagPayload struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type DocReferencePayload struct {
	Raw        string `json:"raw"`
	Label      string `json:"label,omitempty"`
	TargetSlug string `json:"target_slug,omitempty"`
	AssetPath  string `json:"asset_path,omitempty"`
}

type DocNoteMetadataPayload struct {
	Links  []DocReferencePayload `json:"links"`
	Embeds []DocReferencePayload `json:"embeds"`
}

type DocNotePayload struct {
	Slug        string                 `json:"slug"`
	Title       string                 `json:"title"`
	Summary     string                 `json:"summary"`
	ContentType string                 `json:"content_type"`
	Published   bool                   `json:"published,omitempty"`
	Order       int                    `json:"order"`
	Group       string                 `json:"group,omitempty"`
	Tags        []DocTagPayload        `json:"tags"`
	Metadata    DocNoteMetadataPayload `json:"metadata"`
	UpdatedAt   string                 `json:"updated_at"`
	ContentURL  string                 `json:"content_url"`
}

type DocVaultsResponse struct {
	Vaults []DocVaultPayload `json:"vaults"`
}

type DocNotesResponse struct {
	Notes []DocNotePayload `json:"notes"`
}

type DocBranchesResponse struct {
	Branches []string `json:"branches"`
}

type DocAdminSessionRequest struct {
	Session SessionPayload `json:"session"`
}

type DocRegisterVaultRequest struct {
	Session DocSessionPayload `json:"session"`
	Branch  string            `json:"branch"`
	Slug    string            `json:"slug,omitempty"`
	Title   string            `json:"title,omitempty"`
}

type DocVaultActionRequest struct {
	Session DocSessionPayload `json:"session"`
}

type DocOverridePublishedRequest struct {
	Session   DocSessionPayload `json:"session"`
	Published bool              `json:"published"`
}

type DocSessionPayload = SessionPayload

func (p SessionPayload) ToDomain() (domain.SessionData, error) {
	id, err := uuid.Parse(p.UserID)
	if err != nil {
		return domain.SessionData{}, err
	}
	token, err := domain.ParseLoginSessionToken(p.Token)
	if err != nil {
		return domain.SessionData{}, err
	}
	return domain.NewSessionData(id, token)
}

func NewDocVaultPayload(vault domain.DocVault, includeAdminFields bool) DocVaultPayload {
	payload := DocVaultPayload{
		Slug:  vault.Slug,
		Title: vault.Title,
	}
	if includeAdminFields {
		payload.Branch = vault.Branch
		payload.LocalPath = vault.LocalPath
		payload.Status = vault.Status
	}
	if vault.LastSyncedAt != nil {
		formatted := vault.LastSyncedAt.Format("2006-01-02T15:04:05Z07:00")
		payload.LastSyncedAt = &formatted
	}
	return payload
}

func NewDocNotePayload(note domain.DocNote, includeAdminFields bool) DocNotePayload {
	tags := make([]DocTagPayload, len(note.Tags))
	for i, tag := range note.Tags {
		tags[i] = DocTagPayload{Slug: tag.Slug, Name: tag.Name}
	}
	payload := DocNotePayload{
		Slug:        note.Slug,
		Title:       note.Title,
		Summary:     note.Summary,
		ContentType: note.ContentType,
		Order:       note.Order,
		Group:       note.Group,
		Tags:        tags,
		Metadata:    newDocMetadataPayload(note.Metadata),
		UpdatedAt:   note.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ContentURL:  "/api/docs/content/" + note.VaultSlug + "/" + note.Slug,
	}
	if includeAdminFields {
		payload.Published = note.Published
	}
	return payload
}

func newDocMetadataPayload(metadata domain.DocNoteMetadata) DocNoteMetadataPayload {
	return DocNoteMetadataPayload{
		Links:  newDocReferencePayloads(metadata.Links),
		Embeds: newDocReferencePayloads(metadata.Embeds),
	}
}

func newDocReferencePayloads(refs []domain.DocReference) []DocReferencePayload {
	payloads := make([]DocReferencePayload, len(refs))
	for i, ref := range refs {
		payloads[i] = DocReferencePayload{
			Raw:        ref.Raw,
			Label:      ref.Label,
			TargetSlug: ref.TargetSlug,
			AssetPath:  ref.AssetPath,
		}
	}
	return payloads
}
