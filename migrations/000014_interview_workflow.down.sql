DROP INDEX IF EXISTS interviews_real_admin_queue_idx;
DROP INDEX IF EXISTS interviews_real_catalog_idx;
DROP INDEX IF EXISTS interviews_interviewer_kind_status_idx;
DROP INDEX IF EXISTS interviews_student_kind_status_idx;

ALTER TABLE interviews DROP CONSTRAINT IF EXISTS interviews_mock_interviewer_check;
ALTER TABLE interviews DROP CONSTRAINT IF EXISTS interviews_status_kind_check;

ALTER TABLE interviews
    DROP COLUMN IF EXISTS cancelled_at,
    DROP COLUMN IF EXISTS catalog_published,
    DROP COLUMN IF EXISTS cancel_reason,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS external_interviewer,
    DROP COLUMN IF EXISTS student_notes,
    DROP COLUMN IF EXISTS position,
    DROP COLUMN IF EXISTS company;

ALTER TABLE interviews ALTER COLUMN status TYPE TEXT USING status::TEXT;
ALTER TABLE interviews ALTER COLUMN outcome TYPE TEXT USING outcome::TEXT;

DROP TYPE interview_status;
CREATE TYPE interview_status AS ENUM ('scheduled', 'in_progress', 'completed', 'cancelled');

DROP TYPE interview_outcome;
CREATE TYPE interview_outcome AS ENUM ('pending', 'passed', 'failed');

ALTER TABLE interviews
    ALTER COLUMN status TYPE interview_status USING 'scheduled'::interview_status,
    ALTER COLUMN outcome TYPE interview_outcome USING 'pending'::interview_outcome,
    ALTER COLUMN interviewer_id SET NOT NULL;

ALTER TABLE interviews DROP CONSTRAINT IF EXISTS interviews_distinct_users;
ALTER TABLE interviews ADD CONSTRAINT interviews_distinct_users CHECK (student_id <> interviewer_id);

CREATE INDEX interviews_student_status_idx ON interviews (student_id, status, scheduled_at DESC);
CREATE INDEX interviews_interviewer_status_idx ON interviews (interviewer_id, status, scheduled_at DESC);
