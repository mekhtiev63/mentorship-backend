DROP INDEX IF EXISTS final_assessments_roast_reviewer_idx;
DROP INDEX IF EXISTS final_assessments_tech_reviewer_idx;
DROP INDEX IF EXISTS final_assessments_roast_status_idx;
DROP INDEX IF EXISTS final_assessments_tech_status_idx;
DROP INDEX IF EXISTS final_assessments_student_uidx;

ALTER TABLE final_assessments DROP CONSTRAINT IF EXISTS final_assessments_roast_requires_tech_completed;

ALTER TABLE final_assessments
    DROP COLUMN IF EXISTS finalist_event_emitted,
    DROP COLUMN IF EXISTS roast_fail_reason,
    DROP COLUMN IF EXISTS tech_fail_reason,
    DROP COLUMN IF EXISTS roast_failed_at,
    DROP COLUMN IF EXISTS tech_failed_at,
    DROP COLUMN IF EXISTS roast_scheduled_at,
    DROP COLUMN IF EXISTS tech_scheduled_at;

ALTER TABLE final_assessments ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;

ALTER TABLE final_assessments ALTER COLUMN tech_status TYPE TEXT USING tech_status::TEXT;
ALTER TABLE final_assessments ALTER COLUMN roast_status TYPE TEXT USING roast_status::TEXT;

DROP TYPE final_check_status;
CREATE TYPE assessment_status AS ENUM ('not_started', 'in_progress', 'passed', 'failed', 'cancelled');

ALTER TABLE final_assessments
    ALTER COLUMN tech_status TYPE assessment_status USING 'not_started'::assessment_status,
    ALTER COLUMN roast_status TYPE assessment_status USING 'not_started'::assessment_status;

CREATE UNIQUE INDEX final_assessments_one_open_per_student_uidx
    ON final_assessments (student_id) WHERE cancelled_at IS NULL;

CREATE INDEX final_assessments_tech_reviewer_idx
    ON final_assessments (tech_reviewer_id) WHERE cancelled_at IS NULL;
