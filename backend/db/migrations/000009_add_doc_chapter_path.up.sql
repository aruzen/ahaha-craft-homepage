ALTER TABLE doc_notes
    ADD COLUMN chapter_path text NOT NULL DEFAULT '';

CREATE INDEX doc_notes_book_tree_idx
    ON doc_notes (vault_slug, note_group, chapter_path, display_order, title);
