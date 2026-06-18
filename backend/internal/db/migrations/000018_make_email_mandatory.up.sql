UPDATE users 
  SET email = 'noemail_' || id || '@dozingo.de' 
  WHERE email IS NULL;

ALTER TABLE users ALTER COLUMN email SET NOT NULL;
