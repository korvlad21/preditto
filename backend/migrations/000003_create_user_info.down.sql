BEGIN;

DROP TRIGGER IF EXISTS user_info_set_updated_at ON user_info;

DROP TABLE user_info;

DROP FUNCTION IF EXISTS set_user_info_updated_at();

COMMIT;
