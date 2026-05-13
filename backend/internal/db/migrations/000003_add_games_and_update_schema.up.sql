-- ============================================
-- Drop lecturers table and related FK on boards
-- ============================================

ALTER TABLE boards DROP COLUMN IF EXISTS lecturer_id;
DROP TRIGGER IF EXISTS set_updated_at ON lecturers;
DROP TABLE IF EXISTS lecturers;

-- ============================================
-- Modify boards: add description column
-- ============================================

ALTER TABLE boards ADD COLUMN description TEXT;

-- ============================================
-- Modify cells: add value and author_id columns
-- ============================================

ALTER TABLE cells ADD COLUMN value INTEGER;
ALTER TABLE cells ADD COLUMN author_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- ============================================
-- Create games table
-- ============================================

CREATE TABLE games (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    board_id   UUID REFERENCES boards(id) ON DELETE SET NULL,
    status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'abandoned')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON games
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================
-- Create game_cells table
-- ============================================

CREATE TABLE game_cells (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id    UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    cell_id    UUID REFERENCES cells(id) ON DELETE SET NULL,
    content    TEXT NOT NULL,
    position   INTEGER NOT NULL,
    is_marked  BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (game_id, position)
);
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON game_cells
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
