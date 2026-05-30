CREATE TABLE in_app_notifications (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_user_id   UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    notification_type   TEXT NOT NULL,
    title               TEXT NOT NULL,
    body                TEXT NOT NULL DEFAULT '',
    payload             JSONB NOT NULL DEFAULT '{}'::jsonb,
    actor_id            UUID REFERENCES users (id) ON DELETE SET NULL,
    reference_type      TEXT,
    reference_id        TEXT,
    source_outbox_id    UUID NOT NULL,
    source_event_name   TEXT NOT NULL,
    occurred_at         TIMESTAMPTZ NOT NULL,
    read_at             TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT in_app_notifications_type_nonempty CHECK (length(trim(notification_type)) > 0),
    CONSTRAINT in_app_notifications_title_nonempty CHECK (length(trim(title)) > 0)
);

CREATE UNIQUE INDEX in_app_notifications_idempotency_uidx
    ON in_app_notifications (source_outbox_id, recipient_user_id, notification_type);

CREATE INDEX in_app_notifications_recipient_created_idx
    ON in_app_notifications (recipient_user_id, created_at DESC);

CREATE INDEX in_app_notifications_recipient_unread_idx
    ON in_app_notifications (recipient_user_id, created_at DESC)
    WHERE read_at IS NULL;

CREATE INDEX in_app_notifications_recipient_type_idx
    ON in_app_notifications (recipient_user_id, notification_type, created_at DESC);

CREATE TABLE notification_outbox_receipts (
    outbox_id    UUID PRIMARY KEY REFERENCES outbox_events (id) ON DELETE CASCADE,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
