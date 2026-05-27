DROP INDEX IF EXISTS idx_boards_title_trgm;
-- Intentionally NOT dropping the pg_trgm extension: other features may
-- depend on it, and CREATE EXTENSION in the up migration is idempotent.
