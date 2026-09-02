ALTER TABLE doc_vaults
    ADD COLUMN source_type varchar(32) NOT NULL DEFAULT 'git_vault'
        CHECK (source_type IN ('git_vault', 'local_upload'));

ALTER TABLE doc_vaults ALTER COLUMN branch DROP NOT NULL;

CREATE TABLE doc_note_overrides
(
    vault_slug    varchar(160) NOT NULL,
    note_slug     varchar(220) NOT NULL,
    title         varchar(255),
    summary       text,
    published     boolean,
    display_order integer,
    note_group    varchar(160),
    tags          jsonb,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (vault_slug, note_slug),
    FOREIGN KEY (vault_slug) REFERENCES doc_vaults (slug) ON DELETE CASCADE
);

CREATE INDEX doc_notes_public_list_idx
    ON doc_notes (published, vault_slug, note_group, display_order, updated_at DESC);
CREATE INDEX doc_note_tags_tag_idx
    ON doc_note_tags (tag_slug, vault_slug, note_slug);
