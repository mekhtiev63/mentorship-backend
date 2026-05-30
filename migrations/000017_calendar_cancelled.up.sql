ALTER TABLE calendar_events
    ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS calendar_events_upcoming_idx
    ON calendar_events (starts_at)
    WHERE deleted_at IS NULL AND cancelled_at IS NULL;
