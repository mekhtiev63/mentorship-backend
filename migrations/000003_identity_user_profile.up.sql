-- Users, roles, buddy assignments, profiles, refresh tokens, idempotency store.

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           CITEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    status          user_status NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    CONSTRAINT users_email_nonempty CHECK (length(trim(email::text)) > 0),
    CONSTRAINT users_password_hash_nonempty CHECK (length(password_hash) > 0)
);

CREATE UNIQUE INDEX users_email_active_uidx
    ON users (email)
    WHERE deleted_at IS NULL;

CREATE INDEX users_status_idx ON users (status) WHERE deleted_at IS NULL;

CREATE TABLE user_roles (
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role        app_role NOT NULL,
    granted_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role)
);

CREATE INDEX user_roles_role_user_idx ON user_roles (role, user_id);

CREATE TABLE buddy_assignments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id  UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    buddy_id    UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    active      BOOLEAN NOT NULL DEFAULT true,
    valid_from  TIMESTAMPTZ NOT NULL DEFAULT now(),
    valid_to    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    CONSTRAINT buddy_assignments_distinct_users CHECK (student_id <> buddy_id),
    CONSTRAINT buddy_assignments_valid_range CHECK (valid_to IS NULL OR valid_to > valid_from)
);

-- Invariant: at most one active assignment per student.
CREATE UNIQUE INDEX buddy_assignments_one_active_per_student_uidx
    ON buddy_assignments (student_id)
    WHERE active = true AND deleted_at IS NULL;

CREATE INDEX buddy_assignments_buddy_active_idx
    ON buddy_assignments (buddy_id)
    WHERE active = true AND deleted_at IS NULL;

CREATE INDEX buddy_assignments_student_history_idx
    ON buddy_assignments (student_id, created_at DESC);

CREATE TABLE profiles (
    user_id         UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    display_name    TEXT NOT NULL DEFAULT '',
    bio             TEXT NOT NULL DEFAULT '',
    avatar_url      TEXT,
    visibility      profile_visibility NOT NULL DEFAULT 'buddies_only',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT refresh_tokens_hash_nonempty CHECK (length(token_hash) > 0)
);

CREATE UNIQUE INDEX refresh_tokens_token_hash_uidx ON refresh_tokens (token_hash);

CREATE INDEX refresh_tokens_user_active_idx
    ON refresh_tokens (user_id, expires_at DESC)
    WHERE revoked_at IS NULL;

-- Generic idempotency for POST /view, /bonus/convert, etc.
CREATE TABLE idempotency_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    scope           TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    request_hash    TEXT,
    response_code   INT,
    response_body   JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL,
    CONSTRAINT idempotency_keys_scope_nonempty CHECK (length(trim(scope)) > 0),
    CONSTRAINT idempotency_keys_key_nonempty CHECK (length(trim(idempotency_key)) > 0)
);

CREATE UNIQUE INDEX idempotency_keys_user_scope_key_uidx
    ON idempotency_keys (user_id, scope, idempotency_key);

CREATE INDEX idempotency_keys_expires_at_idx ON idempotency_keys (expires_at);
