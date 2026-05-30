DELETE FROM achievement_definitions
WHERE code IN (
    'first_material_view',
    'first_block_approved',
    'blocks_approved_3',
    'blocks_approved_5',
    'program_completed'
);
