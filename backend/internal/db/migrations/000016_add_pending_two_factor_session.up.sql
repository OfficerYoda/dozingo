ALTER TABLE sessions
    ADD COLUMN is_two_fa_pending boolean NOT NULL DEFAULT false;
