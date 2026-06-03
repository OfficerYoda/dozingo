-- Hashing tokens with SHA-256 is irreversible: there is no way to recover
-- the plaintext from a stored digest. This down migration is a no-op.
-- Rolling back the schema would not invalidate the application contract
-- (the column still holds opaque strings); revert the application code to
-- restore the plaintext-storage behavior.
SELECT 1;
