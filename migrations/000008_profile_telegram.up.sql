ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS telegram_username TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS profiles_telegram_username_uidx
    ON profiles (lower(telegram_username))
    WHERE telegram_username IS NOT NULL;
