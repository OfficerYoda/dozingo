ALTER TABLE games ADD COLUMN bingo_count integer NOT NULL DEFAULT 0;
ALTER TABLE games ADD CONSTRAINT chk_bingo_count CHECK (bingo_count >= 0);

UPDATE games SET status = 'active' WHERE status = 'completed';

ALTER TYPE game_status RENAME TO game_status_old;
CREATE TYPE game_status AS ENUM ('active', 'abandoned');
ALTER TABLE games
    ALTER COLUMN status DROP DEFAULT;
ALTER TABLE games
    ALTER COLUMN status TYPE game_status
    USING status::text::game_status;
ALTER TABLE games
    ALTER COLUMN status SET DEFAULT 'active';
DROP TYPE game_status_old;
