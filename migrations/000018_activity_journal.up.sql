-- Activity journal: subject, type, outbox idempotency, occurred_at.

ALTER TABLE activity_events
    ADD COLUMN IF NOT EXISTS subject_user_id UUID REFERENCES users (id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS activity_type TEXT,
    ADD COLUMN IF NOT EXISTS source_outbox_id UUID,
    ADD COLUMN IF NOT EXISTS source_event_name TEXT,
    ADD COLUMN IF NOT EXISTS occurred_at TIMESTAMPTZ;

UPDATE activity_events
SET
    subject_user_id = COALESCE(subject_user_id, actor_id),
    activity_type = COALESCE(activity_type, verb),
    source_event_name = COALESCE(source_event_name, verb),
    occurred_at = COALESCE(occurred_at, created_at)
WHERE subject_user_id IS NULL OR activity_type IS NULL OR occurred_at IS NULL;

DELETE FROM activity_events WHERE subject_user_id IS NULL;

ALTER TABLE activity_events
    ALTER COLUMN subject_user_id SET NOT NULL,
    ALTER COLUMN activity_type SET NOT NULL,
    ALTER COLUMN source_event_name SET NOT NULL,
    ALTER COLUMN occurred_at SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS activity_events_source_outbox_uidx
    ON activity_events (source_outbox_id)
    WHERE source_outbox_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS activity_events_subject_occurred_idx
    ON activity_events (subject_user_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS activity_events_type_subject_idx
    ON activity_events (activity_type, subject_user_id, occurred_at DESC);
