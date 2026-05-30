DROP INDEX IF EXISTS materials_block_active_idx;
DROP INDEX IF EXISTS materials_student_visible_idx;
DROP INDEX IF EXISTS materials_block_sort_order_active_uidx;

DROP INDEX IF EXISTS roadmap_blocks_admin_list_idx;
DROP INDEX IF EXISTS roadmap_blocks_student_catalog_idx;

DROP INDEX IF EXISTS roadmap_blocks_sort_order_active_uidx;
CREATE UNIQUE INDEX roadmap_blocks_sort_order_active_uidx
    ON roadmap_blocks (sort_order)
    WHERE deleted_at IS NULL;

CREATE INDEX roadmap_blocks_published_list_idx
    ON roadmap_blocks (sort_order)
    WHERE status = 'published' AND deleted_at IS NULL;

CREATE INDEX materials_block_sort_order_active_uidx
    ON materials (block_id, sort_order)
    WHERE deleted_at IS NULL;

CREATE INDEX materials_block_active_idx
    ON materials (block_id)
    WHERE deleted_at IS NULL;

ALTER TABLE materials DROP COLUMN IF EXISTS is_active;
ALTER TABLE roadmap_blocks DROP COLUMN IF EXISTS is_active;
