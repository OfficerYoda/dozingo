CREATE TRIGGER set_updated_at
BEFORE UPDATE ON user_two_factors
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON recovery_codes
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
