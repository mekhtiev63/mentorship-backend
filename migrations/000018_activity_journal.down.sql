DROP INDEX IF EXISTS activity_events_type_subject_idx;
DROP INDEX IF EXISTS activity_events_subject_occurred_idx;
DROP INDEX IF EXISTS activity_events_source_outbox_uidx;

ALTER TABLE activity_events
    DROP COLUMN IF EXISTS occurred_at,
    DROP COLUMN IF EXISTS source_event_name,
    DROP COLUMN IF EXISTS source_outbox_id,
    DROP COLUMN IF EXISTS activity_type,
    DROP COLUMN IF EXISTS subject_user_id;
