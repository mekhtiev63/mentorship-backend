ALTER TABLE roadmap_blocks
    ADD COLUMN IF NOT EXISTS expected_skills TEXT[] NOT NULL DEFAULT '{}';
