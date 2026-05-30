CREATE TABLE user_preferences (
    user_id     UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    active_role app_role,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
