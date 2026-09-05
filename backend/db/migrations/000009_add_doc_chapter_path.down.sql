DROP INDEX IF EXISTS doc_notes_book_tree_idx;

ALTER TABLE doc_notes
    DROP COLUMN chapter_path;
