-- Roadmap catalog and student progress.

CREATE TABLE roadmap_blocks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sort_order      INT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    status          roadmap_block_status NOT NULL DEFAULT 'draft',
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    CONSTRAINT roadmap_blocks_title_nonempty CHECK (length(trim(title)) > 0),
    CONSTRAINT roadmap_blocks_sort_order_positive CHECK (sort_order > 0),
    CONSTRAINT roadmap_blocks_published_consistency CHECK (
        (status = 'draft' AND published_at IS NULL)
        OR (status = 'published' AND published_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX roadmap_blocks_sort_order_active_uidx
    ON roadmap_blocks (sort_order)
    WHERE deleted_at IS NULL;

CREATE INDEX roadmap_blocks_published_list_idx
    ON roadmap_blocks (sort_order)
    WHERE status = 'published' AND deleted_at IS NULL;

CREATE TABLE materials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    block_id        UUID NOT NULL REFERENCES roadmap_blocks (id) ON DELETE RESTRICT,
    sort_order      INT NOT NULL,
    title           TEXT NOT NULL,
    material_type   material_type NOT NULL,
    url             TEXT NOT NULL,
    required        BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    CONSTRAINT materials_title_nonempty CHECK (length(trim(title)) > 0),
    CONSTRAINT materials_url_nonempty CHECK (length(trim(url)) > 0),
    CONSTRAINT materials_sort_order_positive CHECK (sort_order > 0)
);

CREATE UNIQUE INDEX materials_block_sort_order_active_uidx
    ON materials (block_id, sort_order)
    WHERE deleted_at IS NULL;

CREATE INDEX materials_block_active_idx
    ON materials (block_id)
    WHERE deleted_at IS NULL;

CREATE TABLE student_block_progress (
    student_id      UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    block_id        UUID NOT NULL REFERENCES roadmap_blocks (id) ON DELETE RESTRICT,
    status          progress_status NOT NULL DEFAULT 'not_started',
    submitted_at    TIMESTAMPTZ,
    approved_by     UUID REFERENCES users (id) ON DELETE SET NULL,
    approved_at     TIMESTAMPTZ,
    rejected_at     TIMESTAMPTZ,
    reject_reason   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (student_id, block_id),
    CONSTRAINT student_block_progress_approval_consistency CHECK (
        (status IN ('approved', 'rejected') AND approved_by IS NOT NULL AND approved_at IS NOT NULL)
        OR (status NOT IN ('approved', 'rejected'))
    )
);

CREATE INDEX student_block_progress_student_status_idx
    ON student_block_progress (student_id, status);

CREATE INDEX student_block_progress_block_awaiting_idx
    ON student_block_progress (block_id, submitted_at)
    WHERE status = 'awaiting_approval';

CREATE TABLE material_views (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id      UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    material_id     UUID NOT NULL REFERENCES materials (id) ON DELETE RESTRICT,
    first_viewed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    idempotency_key TEXT,
    CONSTRAINT material_views_idempotency_key_nonempty CHECK (
        idempotency_key IS NULL OR length(trim(idempotency_key)) > 0
    )
);

-- Invariant: one view record per student and material.
CREATE UNIQUE INDEX material_views_student_material_uidx
    ON material_views (student_id, material_id);

CREATE UNIQUE INDEX material_views_idempotency_uidx
    ON material_views (student_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX material_views_material_idx ON material_views (material_id);
