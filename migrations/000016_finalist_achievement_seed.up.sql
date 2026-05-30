-- Achievement integration: grant on final_check.both_completed (processed by achievement outbox worker when wired).
INSERT INTO achievement_definitions (code, title, description, rule_json)
VALUES (
    'finalist',
    'Финалист',
    'Successfully completed final technical and roast assessments',
    '{"type":"first_event","on":"final_check.both_completed","metric":"final_check_completed","eq":1}'::jsonb
)
ON CONFLICT (code) DO NOTHING;
