-- Restoring NOT NULL on users.email is destructive: any user inserted while
-- email was nullable that doesn't have an email now violates the constraint.
-- We delete those users (and via ON DELETE CASCADE all their dependent rows)
-- so this rollback is runnable without manual cleanup.
DELETE FROM users WHERE email IS NULL;
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
