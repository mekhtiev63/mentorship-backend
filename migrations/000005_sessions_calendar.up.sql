-- One-on-one, calendar, interviews, final assessments.

CREATE TABLE one_on_one_requests (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id          UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    buddy_id            UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status              one_on_one_status NOT NULL DEFAULT 'pending',
    message             TEXT NOT NULL DEFAULT '',
    preferred_slots     JSONB NOT NULL DEFAULT '[]'::jsonb,
    calendar_event_id   UUID,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    cancelled_at        TIMESTAMPTZ,
    CONSTRAINT one_on_one_requests_distinct_users CHECK (student_id <> buddy_id)
);

CREATE INDEX one_on_one_requests_student_status_idx
    ON one_on_one_requests (student_id, status, created_at DESC);

CREATE INDEX one_on_one_requests_buddy_status_idx
    ON one_on_one_requests (buddy_id, status, created_at DESC);

CREATE TABLE calendar_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id    UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    starts_at       TIMESTAMPTZ NOT NULL,
    ends_at         TIMESTAMPTZ NOT NULL,
    related_type    calendar_related_type NOT NULL DEFAULT 'other',
    related_id      UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    CONSTRAINT calendar_events_time_range CHECK (ends_at > starts_at),
    CONSTRAINT calendar_events_title_nonempty CHECK (length(trim(title)) > 0)
);

CREATE INDEX calendar_events_organizer_range_idx
    ON calendar_events (organizer_id, starts_at)
    WHERE deleted_at IS NULL;

CREATE INDEX calendar_events_related_idx
    ON calendar_events (related_type, related_id)
    WHERE deleted_at IS NULL;

CREATE TABLE calendar_event_attendees (
    event_id    UUID NOT NULL REFERENCES calendar_events (id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (event_id, user_id)
);

CREATE INDEX calendar_event_attendees_user_idx ON calendar_event_attendees (user_id);

ALTER TABLE one_on_one_requests
    ADD CONSTRAINT one_on_one_requests_calendar_event_fk
    FOREIGN KEY (calendar_event_id) REFERENCES calendar_events (id) ON DELETE SET NULL;

CREATE TABLE interviews (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id      UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    interviewer_id  UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    kind            interview_kind NOT NULL,
    status          interview_status NOT NULL DEFAULT 'scheduled',
    scheduled_at    TIMESTAMPTZ,
    feedback        TEXT,
    outcome         interview_outcome NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ,
    CONSTRAINT interviews_distinct_users CHECK (student_id <> interviewer_id)
);

CREATE INDEX interviews_student_status_idx
    ON interviews (student_id, status, scheduled_at DESC);

CREATE INDEX interviews_interviewer_status_idx
    ON interviews (interviewer_id, status, scheduled_at DESC);

CREATE TABLE final_assessments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id          UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    tech_status         assessment_status NOT NULL DEFAULT 'not_started',
    roast_status        assessment_status NOT NULL DEFAULT 'not_started',
    tech_reviewer_id    UUID REFERENCES users (id) ON DELETE SET NULL,
    roast_reviewer_id   UUID REFERENCES users (id) ON DELETE SET NULL,
    tech_feedback       TEXT,
    roast_feedback      TEXT,
    tech_completed_at   TIMESTAMPTZ,
    roast_completed_at  TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    cancelled_at        TIMESTAMPTZ,
    CONSTRAINT final_assessments_roast_after_tech CHECK (
        roast_status IN ('not_started', 'cancelled')
        OR tech_status = 'passed'
        OR tech_status = 'cancelled'
    )
);

-- Invariant: one non-cancelled assessment per student.
CREATE UNIQUE INDEX final_assessments_one_open_per_student_uidx
    ON final_assessments (student_id)
    WHERE cancelled_at IS NULL;

CREATE INDEX final_assessments_tech_reviewer_idx
    ON final_assessments (tech_reviewer_id)
    WHERE cancelled_at IS NULL;
