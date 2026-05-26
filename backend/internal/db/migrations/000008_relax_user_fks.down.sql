-- WARNING: this rollback re-introduces the NOT NULL constraint on
-- boards.author_id. It will fail if any board currently has author_id IS NULL
-- (i.e. its author was deleted while running on the new schema). Either
-- assign such boards to a backfill user or hard-delete them before running
-- this down migration.
ALTER TABLE cells DROP CONSTRAINT cells_author_id_fkey;
ALTER TABLE cells
    ADD CONSTRAINT cells_author_id_fkey
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE boards DROP CONSTRAINT boards_author_id_fkey;
ALTER TABLE boards
    ADD CONSTRAINT bingo_boards_author_id_fkey
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE boards ALTER COLUMN author_id SET NOT NULL;
