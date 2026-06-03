-- Hash any plaintext tokens still stored in sessions and verification_tokens
-- so that all rows match the SHA-256 hex digest format expected by the
-- application after this deploy.
--
-- Idempotency / safety: the application only ever writes 64-char lowercase
-- hex digests from this point on, so we restrict the UPDATE to rows whose
-- token does not already look like a SHA-256 hex digest. Re-running this
-- migration is therefore a no-op.

UPDATE sessions
SET token = encode(sha256(token::bytea), 'hex')
WHERE token !~ '^[0-9a-f]{64}$';

UPDATE verification_tokens
SET token = encode(sha256(token::bytea), 'hex')
WHERE token !~ '^[0-9a-f]{64}$';
