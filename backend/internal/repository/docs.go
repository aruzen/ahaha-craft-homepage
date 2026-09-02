package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DocRepository struct {
	db *pgxpool.Pool
}

func NewDocRepository(db *pgxpool.Pool) *DocRepository {
	return &DocRepository{db: db}
}

func (r *DocRepository) ListVaults(ctx context.Context, publicOnly bool) ([]domain.DocVault, error) {
	query := `
		SELECT slug, title, branch, local_path, status, last_synced_at
		FROM doc_vaults
	`
	if publicOnly {
		query += " WHERE status = 'active'"
	}
	query += " ORDER BY title, slug"

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vaults []domain.DocVault
	for rows.Next() {
		vault, err := scanDocVault(rows)
		if err != nil {
			return nil, err
		}
		vaults = append(vaults, vault)
	}
	return vaults, rows.Err()
}

func (r *DocRepository) GetVault(ctx context.Context, slug string) (domain.DocVault, error) {
	row := r.db.QueryRow(ctx, `
		SELECT slug, title, branch, local_path, status, last_synced_at
		FROM doc_vaults
		WHERE slug = $1
	`, slug)
	return scanDocVault(row)
}

func (r *DocRepository) UpsertVault(ctx context.Context, vault domain.DocVault) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO doc_vaults (slug, title, branch, local_path, status, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (slug) DO UPDATE SET
			title = EXCLUDED.title,
			branch = EXCLUDED.branch,
			local_path = EXCLUDED.local_path,
			status = EXCLUDED.status,
			updated_at = NOW()
	`, vault.Slug, vault.Title, vault.Branch, vault.LocalPath, vault.Status)
	return err
}

func (r *DocRepository) DisableVault(ctx context.Context, slug string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE doc_vaults
		SET status = 'disabled', updated_at = NOW()
		WHERE slug = $1
	`, slug)
	return err
}

func (r *DocRepository) MarkVaultSynced(ctx context.Context, slug string, syncedAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		UPDATE doc_vaults
		SET last_synced_at = $2, updated_at = NOW()
		WHERE slug = $1
	`, slug, syncedAt)
	return err
}

func (r *DocRepository) ReplaceScan(ctx context.Context, vaultSlug string, notes []domain.DocNote, assets []domain.DocAsset) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM doc_assets WHERE vault_slug = $1`, vaultSlug); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM doc_note_tags WHERE vault_slug = $1`, vaultSlug); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM doc_tags WHERE vault_slug = $1`, vaultSlug); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM doc_notes WHERE vault_slug = $1`, vaultSlug); err != nil {
		return err
	}

	for _, note := range notes {
		metadataJSON, err := json.Marshal(note.Metadata)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO doc_notes
				(vault_slug, slug, title, summary, source_path, content_type, published, display_order, note_group, metadata, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, vaultSlug, note.Slug, note.Title, note.Summary, note.SourcePath, note.ContentType, note.Published, note.Order, note.Group, metadataJSON, note.UpdatedAt); err != nil {
			return err
		}
		for _, tag := range note.Tags {
			if _, err := tx.Exec(ctx, `
				INSERT INTO doc_tags (vault_slug, slug, name)
				VALUES ($1, $2, $3)
				ON CONFLICT (vault_slug, slug) DO UPDATE SET name = EXCLUDED.name
			`, vaultSlug, tag.Slug, tag.Name); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO doc_note_tags (vault_slug, note_slug, tag_slug)
				VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, vaultSlug, note.Slug, tag.Slug); err != nil {
				return err
			}
		}
	}

	for _, asset := range assets {
		if _, err := tx.Exec(ctx, `
			INSERT INTO doc_assets (vault_slug, asset_path, content_type, size_bytes, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`, vaultSlug, asset.AssetPath, asset.ContentType, asset.SizeBytes, asset.UpdatedAt); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *DocRepository) ListNotes(ctx context.Context, vaultSlug, tag, group string, publicOnly bool) ([]domain.DocNote, error) {
	args := []any{vaultSlug}
	query := `
		SELECT n.vault_slug, n.slug, n.title, n.summary, n.source_path, n.content_type, n.published,
		       n.display_order, n.note_group, n.metadata, n.updated_at
		FROM doc_notes n
	`
	if tag != "" {
		query += ` JOIN doc_note_tags nt ON nt.vault_slug = n.vault_slug AND nt.note_slug = n.slug`
		args = append(args, tag)
	}
	query += ` WHERE n.vault_slug = $1`
	if tag != "" {
		query += ` AND nt.tag_slug = $2`
	}
	if group != "" {
		args = append(args, group)
		query += fmt.Sprintf(` AND n.note_group = $%d`, len(args))
	}
	if publicOnly {
		query += ` AND n.published = true`
	}
	query += ` ORDER BY n.display_order, n.title, n.slug`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []domain.DocNote
	for rows.Next() {
		note, err := scanDocNote(rows)
		if err != nil {
			return nil, err
		}
		tags, err := r.listNoteTags(ctx, note.VaultSlug, note.Slug)
		if err != nil {
			return nil, err
		}
		note.Tags = tags
		notes = append(notes, note)
	}
	return notes, rows.Err()
}

func (r *DocRepository) GetNote(ctx context.Context, vaultSlug, noteSlug string, publicOnly bool) (domain.DocNote, error) {
	query := `
		SELECT vault_slug, slug, title, summary, source_path, content_type, published,
		       display_order, note_group, metadata, updated_at
		FROM doc_notes
		WHERE vault_slug = $1 AND slug = $2
	`
	if publicOnly {
		query += ` AND published = true`
	}
	note, err := scanDocNote(r.db.QueryRow(ctx, query, vaultSlug, noteSlug))
	if err != nil {
		return domain.DocNote{}, err
	}
	tags, err := r.listNoteTags(ctx, note.VaultSlug, note.Slug)
	if err != nil {
		return domain.DocNote{}, err
	}
	note.Tags = tags
	return note, nil
}

func (r *DocRepository) OverrideNotePublished(ctx context.Context, vaultSlug, noteSlug string, published bool) error {
	_, err := r.db.Exec(ctx, `
		UPDATE doc_notes
		SET published = $3, updated_at = NOW()
		WHERE vault_slug = $1 AND slug = $2
	`, vaultSlug, noteSlug, published)
	return err
}

func (r *DocRepository) listNoteTags(ctx context.Context, vaultSlug, noteSlug string) ([]domain.DocTag, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.slug, t.name
		FROM doc_tags t
		JOIN doc_note_tags nt ON nt.vault_slug = t.vault_slug AND nt.tag_slug = t.slug
		WHERE nt.vault_slug = $1 AND nt.note_slug = $2
		ORDER BY t.name
	`, vaultSlug, noteSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []domain.DocTag
	for rows.Next() {
		var tag domain.DocTag
		if err := rows.Scan(&tag.Slug, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

type docRowScanner interface {
	Scan(dest ...any) error
}

func scanDocVault(row docRowScanner) (domain.DocVault, error) {
	var vault domain.DocVault
	var lastSyncedAt *time.Time
	err := row.Scan(&vault.Slug, &vault.Title, &vault.Branch, &vault.LocalPath, &vault.Status, &lastSyncedAt)
	vault.LastSyncedAt = lastSyncedAt
	return vault, err
}

func scanDocNote(row docRowScanner) (domain.DocNote, error) {
	var note domain.DocNote
	var metadataJSON []byte
	err := row.Scan(
		&note.VaultSlug,
		&note.Slug,
		&note.Title,
		&note.Summary,
		&note.SourcePath,
		&note.ContentType,
		&note.Published,
		&note.Order,
		&note.Group,
		&metadataJSON,
		&note.UpdatedAt,
	)
	if err != nil {
		return domain.DocNote{}, err
	}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &note.Metadata); err != nil {
			return domain.DocNote{}, err
		}
	}
	return note, nil
}
