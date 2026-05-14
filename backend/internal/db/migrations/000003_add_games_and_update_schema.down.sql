-- ============================================
-- Drop game_cells and games tables
-- ============================================

DROP TRIGGER IF EXISTS set_updated_at ON game_cells;
DROP TABLE IF EXISTS game_cells;

DROP TRIGGER IF EXISTS set_updated_at ON games;
DROP TABLE IF EXISTS games;

-- ============================================
-- Revert cells: drop value and author_id columns
-- ============================================

ALTER TABLE cells DROP COLUMN IF EXISTS author_id;
ALTER TABLE cells DROP COLUMN IF EXISTS value;

-- ============================================
-- Revert boards: drop description column
-- ============================================

ALTER TABLE boards DROP COLUMN IF EXISTS description;

-- ============================================
-- Recreate lecturers table and add lecturer_id back to boards
-- ============================================

CREATE TABLE lecturers (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    slug       TEXT UNIQUE NOT NULL CHECK (slug ~ '^[a-z0-9-]+$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON lecturers
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE boards ADD COLUMN lecturer_id UUID REFERENCES lecturers(id) ON DELETE SET NULL;
