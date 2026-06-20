ALTER TABLE games DROP CONSTRAINT IF EXISTS chk_bingo_count;
ALTER TABLE games DROP COLUMN IF EXISTS bingo_count;

ALTER TYPE game_status RENAME TO game_status_old;
CREATE TYPE game_status AS ENUM ('active', 'completed', 'abandoned');
ALTER TABLE games
    ALTER COLUMN status DROP DEFAULT;
ALTER TABLE games
    ALTER COLUMN status TYPE game_status
    USING status::text::game_status;
ALTER TABLE games
    ALTER COLUMN status SET DEFAULT 'active';
DROP TYPE game_status_old;
