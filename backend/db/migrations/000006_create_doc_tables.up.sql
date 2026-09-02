CREATE TABLE doc_vaults
(
    slug           varchar(160) PRIMARY KEY,
    title          varchar(255) NOT NULL,
    branch         varchar(255) NOT NULL UNIQUE,
    local_path     text         NOT NULL,
    status         varchar(32)  NOT NULL DEFAULT 'active',
    last_synced_at TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE doc_notes
(
    vault_slug   varchar(160) NOT NULL REFERENCES doc_vaults (slug) ON DELETE CASCADE,
    slug         varchar(220) NOT NULL,
    title        varchar(255) NOT NULL,
    summary      text         NOT NULL DEFAULT '',
    source_path  text         NOT NULL,
    content_type varchar(32)  NOT NULL,
    published    boolean      NOT NULL DEFAULT false,
    display_order integer     NOT NULL DEFAULT 0,
    note_group   varchar(160) NOT NULL DEFAULT '',
    metadata     jsonb        NOT NULL DEFAULT '{}'::jsonb,
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (vault_slug, slug)
);

CREATE TABLE doc_tags
(
    vault_slug varchar(160) NOT NULL REFERENCES doc_vaults (slug) ON DELETE CASCADE,
    slug       varchar(160) NOT NULL,
    name       varchar(160) NOT NULL,
    PRIMARY KEY (vault_slug, slug)
);

CREATE TABLE doc_note_tags
(
    vault_slug varchar(160) NOT NULL,
    note_slug  varchar(220) NOT NULL,
    tag_slug   varchar(160) NOT NULL,
    PRIMARY KEY (vault_slug, note_slug, tag_slug),
    FOREIGN KEY (vault_slug, note_slug) REFERENCES doc_notes (vault_slug, slug) ON DELETE CASCADE,
    FOREIGN KEY (vault_slug, tag_slug) REFERENCES doc_tags (vault_slug, slug) ON DELETE CASCADE
);

CREATE TABLE doc_assets
(
    vault_slug   varchar(160) NOT NULL REFERENCES doc_vaults (slug) ON DELETE CASCADE,
    asset_path   text         NOT NULL,
    content_type varchar(120) NOT NULL,
    size_bytes   bigint       NOT NULL,
    updated_at   TIMESTAMPTZ  NOT NULL,
    PRIMARY KEY (vault_slug, asset_path)
);
