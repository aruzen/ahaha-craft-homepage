package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"backend/internal/domain"

	"github.com/jackc/pgx/v5"
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
		SELECT slug, title, COALESCE(branch, ''), local_path, status, last_synced_at, source_type
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
		SELECT slug, title, COALESCE(branch, ''), local_path, status, last_synced_at, source_type
		FROM doc_vaults
		WHERE slug = $1
	`, slug)
	return scanDocVault(row)
}

func (r *DocRepository) UpsertVault(ctx context.Context, vault domain.DocVault) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO doc_vaults (slug, title, branch, local_path, status, source_type, updated_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, NOW())
		ON CONFLICT (slug) DO UPDATE SET
			title = EXCLUDED.title,
			branch = EXCLUDED.branch,
			local_path = EXCLUDED.local_path,
			status = EXCLUDED.status,
			source_type = EXCLUDED.source_type,
			updated_at = NOW()
	`, vault.Slug, vault.Title, vault.Branch, vault.LocalPath, vault.Status, vault.SourceType)
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
		note, err = applyNoteOverride(ctx, tx, vaultSlug, note)
		if err != nil {
			return err
		}
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

func (r *DocRepository) ListPublishedToys(ctx context.Context) ([]domain.DocToy, error) {
	rows, err := r.db.Query(ctx, `
		SELECT v.slug, v.title, COALESCE(v.branch, ''), v.local_path, v.status, v.last_synced_at, v.source_type,
		       n.vault_slug, n.slug, n.title, n.summary, n.source_path, n.content_type, n.published,
		       n.display_order, n.note_group, n.metadata, n.updated_at
		FROM doc_notes n
		JOIN doc_vaults v ON v.slug = n.vault_slug
		WHERE v.status = 'active' AND n.published = true
		ORDER BY n.updated_at DESC, n.display_order, n.title
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var toys []domain.DocToy
	for rows.Next() {
		var toy domain.DocToy
		var lastSynced *time.Time
		var metadataJSON []byte
		if err := rows.Scan(
			&toy.Vault.Slug, &toy.Vault.Title, &toy.Vault.Branch, &toy.Vault.LocalPath, &toy.Vault.Status, &lastSynced, &toy.Vault.SourceType,
			&toy.Note.VaultSlug, &toy.Note.Slug, &toy.Note.Title, &toy.Note.Summary, &toy.Note.SourcePath,
			&toy.Note.ContentType, &toy.Note.Published, &toy.Note.Order, &toy.Note.Group, &metadataJSON, &toy.Note.UpdatedAt,
		); err != nil {
			return nil, err
		}
		toy.Vault.LastSyncedAt = lastSynced
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &toy.Note.Metadata); err != nil {
				return nil, err
			}
		}
		toys = append(toys, toy)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	tagRows, err := r.db.Query(ctx, `
		SELECT nt.vault_slug, nt.note_slug, t.slug, t.name
		FROM doc_note_tags nt JOIN doc_tags t
		  ON t.vault_slug = nt.vault_slug AND t.slug = nt.tag_slug
		JOIN doc_notes n ON n.vault_slug = nt.vault_slug AND n.slug = nt.note_slug
		JOIN doc_vaults v ON v.slug = n.vault_slug
		WHERE v.status = 'active' AND n.published = true
		ORDER BY t.name
	`)
	if err != nil {
		return nil, err
	}
	defer tagRows.Close()
	tags := map[string][]domain.DocTag{}
	for tagRows.Next() {
		var vaultSlug, noteSlug string
		var tag domain.DocTag
		if err := tagRows.Scan(&vaultSlug, &noteSlug, &tag.Slug, &tag.Name); err != nil {
			return nil, err
		}
		key := vaultSlug + "\x00" + noteSlug
		tags[key] = append(tags[key], tag)
	}
	for i := range toys {
		toys[i].Note.Tags = tags[toys[i].Note.VaultSlug+"\x00"+toys[i].Note.Slug]
	}
	return toys, tagRows.Err()
}

func (r *DocRepository) UpsertNoteOverride(ctx context.Context, vaultSlug, noteSlug string, override domain.DocNoteOverride) error {
	var tagsJSON *string
	if override.Tags != nil {
		encoded, err := json.Marshal(*override.Tags)
		if err != nil {
			return err
		}
		value := string(encoded)
		tagsJSON = &value
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO doc_note_overrides
			(vault_slug, note_slug, title, summary, published, display_order, note_group, tags, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
		ON CONFLICT (vault_slug, note_slug) DO UPDATE SET
			title=EXCLUDED.title, summary=EXCLUDED.summary, published=EXCLUDED.published,
			display_order=EXCLUDED.display_order, note_group=EXCLUDED.note_group,
			tags=EXCLUDED.tags, updated_at=NOW()
	`, vaultSlug, noteSlug, override.Title, override.Summary, override.Published, override.Order, override.Group, tagsJSON)
	return err
}

func applyNoteOverride(ctx context.Context, tx pgx.Tx, vaultSlug string, note domain.DocNote) (domain.DocNote, error) {
	var title, summary, group *string
	var published *bool
	var order *int
	var tagsJSON []byte
	err := tx.QueryRow(ctx, `SELECT title, summary, published, display_order, note_group, tags
		FROM doc_note_overrides WHERE vault_slug=$1 AND note_slug=$2`, vaultSlug, note.Slug).
		Scan(&title, &summary, &published, &order, &group, &tagsJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return note, nil
	}
	if err != nil {
		return note, err
	}
	if title != nil {
		note.Title = *title
	}
	if summary != nil {
		note.Summary = *summary
	}
	if published != nil {
		note.Published = *published
	}
	if order != nil {
		note.Order = *order
	}
	if group != nil {
		note.Group = *group
	}
	if tagsJSON != nil {
		if err := json.Unmarshal(tagsJSON, &note.Tags); err != nil {
			return note, err
		}
	}
	return note, nil
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
	err := row.Scan(&vault.Slug, &vault.Title, &vault.Branch, &vault.LocalPath, &vault.Status, &lastSyncedAt, &vault.SourceType)
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
