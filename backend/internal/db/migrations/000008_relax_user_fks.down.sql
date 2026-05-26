-- Restore the cells.author_id FK action to CASCADE.
ALTER TABLE cells DROP CONSTRAINT cells_author_id_fkey;
ALTER TABLE cells
    ADD CONSTRAINT cells_author_id_fkey
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE;

-- Restoring NOT NULL on boards.author_id is destructive: any board whose
-- author was deleted while running on the new schema now has author_id IS
-- NULL. Drop those boards (and via ON DELETE CASCADE on cells/votes their
-- dependent rows) so the rollback is runnable without manual cleanup.
DELETE FROM boards WHERE author_id IS NULL;
ALTER TABLE boards DROP CONSTRAINT boards_author_id_fkey;
ALTER TABLE boards
    ADD CONSTRAINT bingo_boards_author_id_fkey
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE boards ALTER COLUMN author_id SET NOT NULL;
