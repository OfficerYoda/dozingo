ALTER TABLE boards DROP CONSTRAINT IF EXISTS bingo_boards_size_check;
ALTER TABLE boards DROP CONSTRAINT IF EXISTS boards_size_check;

ALTER TABLE boards
    ADD CONSTRAINT boards_size_check CHECK (size >= 4 AND size <= 6);
