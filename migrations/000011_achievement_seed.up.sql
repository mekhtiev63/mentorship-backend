INSERT INTO achievement_definitions (code, title, description, rule_json)
VALUES
    (
        'first_material_view',
        'First step',
        'Viewed your first learning material',
        '{"type":"first_event","on":"progress.material.viewed","metric":"material_views_count","eq":1}'::jsonb
    ),
    (
        'first_block_approved',
        'First block done',
        'Completed your first roadmap block',
        '{"type":"threshold","on":"progress.block.approved","metric":"approved_blocks_count","gte":1}'::jsonb
    ),
    (
        'blocks_approved_3',
        'Three blocks',
        'Completed three roadmap blocks',
        '{"type":"threshold","on":"progress.block.approved","metric":"approved_blocks_count","gte":3}'::jsonb
    ),
    (
        'blocks_approved_5',
        'Five blocks',
        'Completed five roadmap blocks',
        '{"type":"threshold","on":"progress.block.approved","metric":"approved_blocks_count","gte":5}'::jsonb
    ),
    (
        'program_completed',
        'Program graduate',
        'Completed the full published roadmap',
        '{"type":"all_published_blocks_approved","on":"progress.block.approved"}'::jsonb
    )
ON CONFLICT (code) DO NOTHING;
