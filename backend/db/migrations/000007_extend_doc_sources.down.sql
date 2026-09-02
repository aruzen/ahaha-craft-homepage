DROP INDEX IF EXISTS doc_note_tags_tag_idx;
DROP INDEX IF EXISTS doc_notes_public_list_idx;
DROP TABLE IF EXISTS doc_note_overrides;
DELETE FROM doc_vaults WHERE source_type = 'local_upload';
ALTER TABLE doc_vaults ALTER COLUMN branch SET NOT NULL;
ALTER TABLE doc_vaults DROP COLUMN source_type;
