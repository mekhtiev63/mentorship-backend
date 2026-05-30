-- Interview workflow: statuses, outcomes, real/mock fields.

ALTER TABLE interviews ALTER COLUMN outcome DROP DEFAULT;
ALTER TABLE interviews ALTER COLUMN outcome TYPE TEXT USING outcome::TEXT;
DROP TYPE interview_outcome;
CREATE TYPE interview_outcome AS ENUM ('offer', 'reject', 'pending', 'no_result');

ALTER TABLE interviews ALTER COLUMN status DROP DEFAULT;
ALTER TABLE interviews ALTER COLUMN status TYPE TEXT USING status::TEXT;
DROP TYPE interview_status;
CREATE TYPE interview_status AS ENUM (
    'submitted',
    'reviewed',
    'scheduled',
    'completed',
    'cancelled'
);

ALTER TABLE interviews
    ALTER COLUMN status TYPE interview_status USING (
        CASE
            WHEN kind = 'real' THEN
                CASE status
                    WHEN 'completed' THEN 'completed'::interview_status
                    WHEN 'cancelled' THEN 'cancelled'::interview_status
                    WHEN 'reviewed' THEN 'reviewed'::interview_status
                    ELSE 'submitted'::interview_status
                END
            ELSE
                CASE status
                    WHEN 'completed' THEN 'completed'::interview_status
                    WHEN 'cancelled' THEN 'cancelled'::interview_status
                    ELSE 'scheduled'::interview_status
                END
        END
    ),
    ALTER COLUMN outcome TYPE interview_outcome USING (
        CASE outcome
            WHEN 'passed' THEN 'offer'::interview_outcome
            WHEN 'failed' THEN 'reject'::interview_outcome
            WHEN 'offer' THEN 'offer'::interview_outcome
            WHEN 'reject' THEN 'reject'::interview_outcome
            WHEN 'no_result' THEN 'no_result'::interview_outcome
            ELSE 'pending'::interview_outcome
        END
    );

ALTER TABLE interviews ALTER COLUMN outcome SET DEFAULT 'pending';

ALTER TABLE interviews ALTER COLUMN interviewer_id DROP NOT NULL;

ALTER TABLE interviews
    ADD COLUMN IF NOT EXISTS company TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS position TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS student_notes TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS external_interviewer TEXT,
    ADD COLUMN IF NOT EXISTS reviewed_by UUID REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS cancel_reason TEXT,
    ADD COLUMN IF NOT EXISTS catalog_published BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;

ALTER TABLE interviews DROP CONSTRAINT IF EXISTS interviews_distinct_users;
ALTER TABLE interviews ADD CONSTRAINT interviews_distinct_users CHECK (
    interviewer_id IS NULL OR student_id <> interviewer_id
);

ALTER TABLE interviews ADD CONSTRAINT interviews_status_kind_check CHECK (
    (kind = 'real' AND status IN ('submitted', 'reviewed', 'completed', 'cancelled'))
    OR (kind = 'mock' AND status IN ('scheduled', 'completed', 'cancelled'))
);

ALTER TABLE interviews ADD CONSTRAINT interviews_mock_interviewer_check CHECK (
    kind <> 'mock' OR interviewer_id IS NOT NULL
);

DROP INDEX IF EXISTS interviews_student_status_idx;
DROP INDEX IF EXISTS interviews_interviewer_status_idx;

CREATE INDEX interviews_student_kind_status_idx
    ON interviews (student_id, kind, status, scheduled_at DESC NULLS LAST);

CREATE INDEX interviews_interviewer_kind_status_idx
    ON interviews (interviewer_id, kind, status, scheduled_at DESC NULLS LAST)
    WHERE interviewer_id IS NOT NULL;

CREATE INDEX interviews_real_catalog_idx
    ON interviews (completed_at DESC NULLS LAST)
    WHERE kind = 'real' AND status = 'completed' AND catalog_published = TRUE;

CREATE INDEX interviews_real_admin_queue_idx
    ON interviews (status, created_at DESC)
    WHERE kind = 'real';
