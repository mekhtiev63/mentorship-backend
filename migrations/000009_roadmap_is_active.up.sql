ALTER TABLE roadmap_blocks
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;

ALTER TABLE materials
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;

DROP INDEX IF EXISTS roadmap_blocks_sort_order_active_uidx;
CREATE UNIQUE INDEX roadmap_blocks_sort_order_active_uidx
    ON roadmap_blocks (sort_order)
    WHERE deleted_at IS NULL AND is_active = true;

DROP INDEX IF EXISTS roadmap_blocks_published_list_idx;
CREATE INDEX roadmap_blocks_student_catalog_idx
    ON roadmap_blocks (sort_order)
    WHERE status = 'published' AND is_active = true AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS roadmap_blocks_admin_list_idx
    ON roadmap_blocks (created_at DESC)
    WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS materials_block_sort_order_active_uidx;
CREATE UNIQUE INDEX materials_block_sort_order_active_uidx
    ON materials (block_id, sort_order)
    WHERE deleted_at IS NULL AND is_active = true;

DROP INDEX IF EXISTS materials_block_active_idx;
CREATE INDEX materials_student_visible_idx
    ON materials (block_id, sort_order)
    WHERE is_active = true AND deleted_at IS NULL;

CREATE INDEX materials_block_active_idx
    ON materials (block_id)
    WHERE deleted_at IS NULL;
