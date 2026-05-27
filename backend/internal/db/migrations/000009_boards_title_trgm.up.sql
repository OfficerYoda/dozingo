CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_boards_title_trgm
    ON boards USING GIN (title gin_trgm_ops);
