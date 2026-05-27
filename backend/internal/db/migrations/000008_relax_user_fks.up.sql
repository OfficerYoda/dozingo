-- boards.author_id: drop NOT NULL and switch FK action from CASCADE to SET NULL
-- so that deleting a user preserves their boards as authored-by-null instead
-- of cascading the deletion through to all their content.
--
-- Note: the FK constraint is still named after the original table
-- (bingo_boards) because migration 000002 renamed the table without renaming
-- the constraint. Rename it here to match the current table name as well.
ALTER TABLE boards ALTER COLUMN author_id DROP NOT NULL;
ALTER TABLE boards DROP CONSTRAINT bingo_boards_author_id_fkey;
ALTER TABLE boards
    ADD CONSTRAINT boards_author_id_fkey
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE SET NULL;

-- cells.author_id is already nullable; relax CASCADE -> SET NULL so 'personal
-- cell' rows survive their author's deletion the same way boards do.
ALTER TABLE cells DROP CONSTRAINT cells_author_id_fkey;
ALTER TABLE cells
    ADD CONSTRAINT cells_author_id_fkey
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE SET NULL;
