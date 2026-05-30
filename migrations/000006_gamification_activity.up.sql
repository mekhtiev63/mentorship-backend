-- Achievements, activity feed, bonus ledger, transactional outbox.

CREATE TABLE achievement_definitions (
    code            TEXT PRIMARY KEY,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    rule_json       JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    CONSTRAINT achievement_definitions_code_nonempty CHECK (length(trim(code)) > 0),
    CONSTRAINT achievement_definitions_title_nonempty CHECK (length(trim(title)) > 0)
);

CREATE INDEX achievement_definitions_active_idx
    ON achievement_definitions (code)
    WHERE deleted_at IS NULL;

CREATE TABLE user_achievements (
    user_id             UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    achievement_code    TEXT NOT NULL REFERENCES achievement_definitions (code) ON DELETE RESTRICT,
    granted_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    source_event_id     UUID NOT NULL,
    PRIMARY KEY (user_id, achievement_code)
);

-- Invariant: idempotent grant per domain event.
CREATE UNIQUE INDEX user_achievements_source_event_uidx
    ON user_achievements (source_event_id, achievement_code);

CREATE INDEX user_achievements_user_granted_idx
    ON user_achievements (user_id, granted_at DESC);

CREATE TABLE activity_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id        UUID REFERENCES users (id) ON DELETE SET NULL,
    verb            TEXT NOT NULL,
    object_type     TEXT NOT NULL,
    object_id       UUID,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT activity_events_verb_nonempty CHECK (length(trim(verb)) > 0),
    CONSTRAINT activity_events_object_type_nonempty CHECK (length(trim(object_type)) > 0)
);

CREATE INDEX activity_events_created_at_idx ON activity_events (created_at DESC);

CREATE INDEX activity_events_actor_created_idx ON activity_events (actor_id, created_at DESC);

CREATE TABLE activity_feed_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    event_id        UUID NOT NULL REFERENCES activity_events (id) ON DELETE CASCADE,
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX activity_feed_items_event_user_uidx
    ON activity_feed_items (event_id, user_id);

CREATE INDEX activity_feed_items_user_unread_idx
    ON activity_feed_items (user_id, created_at DESC)
    WHERE read_at IS NULL;

CREATE TABLE bonus_accounts (
    user_id         UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    balance         BIGINT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT bonus_accounts_balance_non_negative CHECK (balance >= 0)
);

CREATE TABLE bonus_transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    amount          BIGINT NOT NULL,
    type            bonus_transaction_type NOT NULL,
    reference       TEXT,
    idempotency_key TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT bonus_transactions_amount_non_zero CHECK (amount <> 0),
    CONSTRAINT bonus_transactions_idempotency_key_nonempty CHECK (
        idempotency_key IS NULL OR length(trim(idempotency_key)) > 0
    )
);

CREATE UNIQUE INDEX bonus_transactions_idempotency_key_uidx
    ON bonus_transactions (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX bonus_transactions_user_created_idx
    ON bonus_transactions (user_id, created_at DESC);

CREATE TABLE outbox_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_name      TEXT NOT NULL,
    payload         JSONB NOT NULL,
    status          outbox_status NOT NULL DEFAULT 'pending',
    attempts        INT NOT NULL DEFAULT 0,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at    TIMESTAMPTZ,
    CONSTRAINT outbox_events_name_nonempty CHECK (length(trim(event_name)) > 0),
    CONSTRAINT outbox_events_attempts_non_negative CHECK (attempts >= 0)
);

CREATE INDEX outbox_events_pending_idx
    ON outbox_events (created_at)
    WHERE status IN ('pending', 'failed');
