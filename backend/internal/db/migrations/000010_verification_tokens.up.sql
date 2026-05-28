CREATE TYPE token_type AS ENUM ('password_reset', 'email_verification');

CREATE TABLE verification_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT UNIQUE NOT NULL,
    type       token_type NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON verification_tokens
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX ON verification_tokens(expires_at);
CREATE INDEX ON verification_tokens(user_id, type);
