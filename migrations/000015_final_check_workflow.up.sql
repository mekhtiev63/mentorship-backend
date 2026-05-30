-- Final check track statuses and assessment workflow.

ALTER TABLE final_assessments DROP CONSTRAINT IF EXISTS final_assessments_roast_after_tech;
DROP INDEX IF EXISTS final_assessments_one_open_per_student_uidx;
DROP INDEX IF EXISTS final_assessments_tech_reviewer_idx;

ALTER TABLE final_assessments ALTER COLUMN tech_status DROP DEFAULT;
ALTER TABLE final_assessments ALTER COLUMN roast_status DROP DEFAULT;
ALTER TABLE final_assessments ALTER COLUMN tech_status TYPE TEXT USING tech_status::TEXT;
ALTER TABLE final_assessments ALTER COLUMN roast_status TYPE TEXT USING roast_status::TEXT;

DROP TYPE IF EXISTS final_check_status;
CREATE TYPE final_check_status AS ENUM (
    'not_available',
    'available',
    'scheduled',
    'completed',
    'failed'
);

ALTER TABLE final_assessments
    ALTER COLUMN tech_status TYPE final_check_status USING (
        CASE tech_status
            WHEN 'passed' THEN 'completed'::final_check_status
            WHEN 'failed' THEN 'failed'::final_check_status
            WHEN 'in_progress' THEN 'scheduled'::final_check_status
            WHEN 'cancelled' THEN 'failed'::final_check_status
            ELSE 'not_available'::final_check_status
        END
    ),
    ALTER COLUMN roast_status TYPE final_check_status USING (
        CASE roast_status
            WHEN 'passed' THEN 'completed'::final_check_status
            WHEN 'failed' THEN 'failed'::final_check_status
            WHEN 'in_progress' THEN 'scheduled'::final_check_status
            WHEN 'cancelled' THEN 'failed'::final_check_status
            ELSE 'not_available'::final_check_status
        END
    );

ALTER TABLE final_assessments
    ALTER COLUMN tech_status SET DEFAULT 'not_available',
    ALTER COLUMN roast_status SET DEFAULT 'not_available';

ALTER TABLE final_assessments
    ADD COLUMN IF NOT EXISTS tech_scheduled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS roast_scheduled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS tech_failed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS roast_failed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS tech_fail_reason TEXT,
    ADD COLUMN IF NOT EXISTS roast_fail_reason TEXT,
    ADD COLUMN IF NOT EXISTS finalist_event_emitted BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE final_assessments DROP COLUMN IF EXISTS cancelled_at;

DROP TYPE IF EXISTS assessment_status;

ALTER TABLE final_assessments ADD CONSTRAINT final_assessments_roast_requires_tech_completed CHECK (
    roast_status = 'not_available' OR tech_status = 'completed'
);

CREATE UNIQUE INDEX IF NOT EXISTS final_assessments_student_uidx ON final_assessments (student_id);

CREATE INDEX final_assessments_tech_status_idx
    ON final_assessments (tech_status, tech_scheduled_at)
    WHERE tech_status IN ('available', 'scheduled');

CREATE INDEX final_assessments_roast_status_idx
    ON final_assessments (roast_status, roast_scheduled_at)
    WHERE roast_status IN ('available', 'scheduled');

CREATE INDEX final_assessments_tech_reviewer_idx
    ON final_assessments (tech_reviewer_id)
    WHERE tech_status = 'scheduled';

CREATE INDEX final_assessments_roast_reviewer_idx
    ON final_assessments (roast_reviewer_id)
    WHERE roast_status = 'scheduled';
