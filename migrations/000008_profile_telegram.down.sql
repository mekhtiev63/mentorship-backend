DROP INDEX IF EXISTS profiles_telegram_username_uidx;

ALTER TABLE profiles
    DROP COLUMN IF EXISTS telegram_username;
