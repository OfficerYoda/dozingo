ALTER TABLE sessions
    ADD COLUMN two_fa_pending boolean NOT NULL DEFAULT false;
