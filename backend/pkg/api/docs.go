package api

import (
	"backend/internal/domain"
	"net/url"
	"path"
	"strings"

	"github.com/google/uuid"
)

type DocVaultPayload struct {
	Slug             string  `json:"slug"`
	Title            string  `json:"title"`
	Branch           string  `json:"branch,omitempty"`
	LocalPath        string  `json:"local_path,omitempty"`
	Status           string  `json:"status,omitempty"`
	LastSyncedAt     *string `json:"last_synced_at,omitempty"`
	SourceType       string  `json:"source_type"`
	DefaultPublished bool    `json:"default_published,omitempty"`
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
	Slug         string                 `json:"slug"`
	Title        string                 `json:"title"`
	Summary      string                 `json:"summary"`
	ContentType  string                 `json:"content_type"`
	Published    bool                   `json:"published,omitempty"`
	Order        int                    `json:"order"`
	Group        string                 `json:"group,omitempty"`
	ChapterPath  []string               `json:"chapter_path"`
	Tags         []DocTagPayload        `json:"tags"`
	Metadata     DocNoteMetadataPayload `json:"metadata"`
	UpdatedAt    string                 `json:"updated_at"`
	ContentURL   string                 `json:"content_url"`
	AssetBaseURL string                 `json:"asset_base_url"`
}

type DocVaultsResponse struct {
	Vaults []DocVaultPayload `json:"vaults"`
}

type DocNotesResponse struct {
	Notes []DocNotePayload `json:"notes"`
}

type DocToyPayload struct {
	Source DocVaultPayload `json:"source"`
	Note   DocNotePayload  `json:"note"`
}

type DocToysResponse struct {
	Toys []DocToyPayload `json:"toys"`
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

type DocDefaultPublishedRequest struct {
	Session          DocSessionPayload `json:"session"`
	DefaultPublished bool              `json:"default_published"`
}

type DocOverridePublishedRequest struct {
	Session   DocSessionPayload `json:"session"`
	Published bool              `json:"published"`
}

type DocUpdateMetadataRequest struct {
	Session   DocSessionPayload `json:"session"`
	Title     *string           `json:"title,omitempty"`
	Summary   *string           `json:"summary,omitempty"`
	Published *bool             `json:"published,omitempty"`
	Order     *int              `json:"order,omitempty"`
	Group     *string           `json:"group,omitempty"`
	Tags      *[]string         `json:"tags,omitempty"`
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
		Slug: vault.Slug, Title: vault.Title, SourceType: vault.SourceType,
	}
	if includeAdminFields {
		payload.Branch = vault.Branch
		payload.LocalPath = vault.LocalPath
		payload.Status = vault.Status
		payload.DefaultPublished = vault.DefaultPublished
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
		Slug:         note.Slug,
		Title:        note.Title,
		Summary:      note.Summary,
		ContentType:  note.ContentType,
		Order:        note.Order,
		Group:        note.Group,
		ChapterPath:  splitChapterPath(note.ChapterPath),
		Tags:         tags,
		Metadata:     newDocMetadataPayload(note.Metadata),
		UpdatedAt:    note.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ContentURL:   "/api/docs/content/" + note.VaultSlug + "/" + note.Slug,
		AssetBaseURL: docAssetBaseURL(note),
	}
	if includeAdminFields {
		payload.Published = note.Published
	}
	return payload
}

func splitChapterPath(value string) []string {
	if value == "" {
		return []string{}
	}
	return strings.Split(value, "/")
}

func docAssetBaseURL(note domain.DocNote) string {
	dir := path.Dir(strings.ReplaceAll(note.SourcePath, "\\", "/"))
	base := "/api/docs/assets/" + url.PathEscape(note.VaultSlug) + "/"
	if dir == "." {
		return base
	}
	parts := strings.Split(dir, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return base + strings.Join(parts, "/") + "/"
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
