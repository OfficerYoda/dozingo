-- Revert games.status from game_status enum to TEXT
ALTER TABLE games ALTER COLUMN status DROP DEFAULT;
ALTER TABLE games
    ALTER COLUMN status TYPE TEXT
    USING status::TEXT;
ALTER TABLE games ALTER COLUMN status SET DEFAULT 'active';
ALTER TABLE games
    ADD CONSTRAINT chk_game_status
    CHECK (status IN ('active', 'completed', 'abandoned'));

DROP TYPE IF EXISTS game_status;

-- Revert users.email_verified_at
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
