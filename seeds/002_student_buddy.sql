-- Development student + buddy (password for both: changeme)

INSERT INTO users (id, email, password_hash, status)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'student@example.com',
    '$2b$12$rovqPB3WrEE3n5JvzeV17..3duKHhBOXMq.3WhjudnPvszRLdi4k6',
    'active'
)
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    status = EXCLUDED.status,
    deleted_at = NULL,
    updated_at = now();

INSERT INTO users (id, email, password_hash, status)
VALUES (
    '00000000-0000-0000-0000-000000000003',
    'buddy@example.com',
    '$2b$12$rovqPB3WrEE3n5JvzeV17..3duKHhBOXMq.3WhjudnPvszRLdi4k6',
    'active'
)
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    status = EXCLUDED.status,
    deleted_at = NULL,
    updated_at = now();

INSERT INTO user_roles (user_id, role)
VALUES ('00000000-0000-0000-0000-000000000002', 'student')
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role)
VALUES ('00000000-0000-0000-0000-000000000003', 'buddy')
ON CONFLICT DO NOTHING;

INSERT INTO profiles (user_id, display_name, bio, visibility)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'Иван Студентов',
    'Go-разработчик в программе менторства.',
    'buddies_only'
)
ON CONFLICT (user_id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    bio = EXCLUDED.bio,
    visibility = EXCLUDED.visibility,
    updated_at = now();

INSERT INTO profiles (user_id, display_name, bio, visibility)
VALUES (
    '00000000-0000-0000-0000-000000000003',
    'Мария Бадди',
    'Senior Go engineer, ментор.',
    'buddies_only'
)
ON CONFLICT (user_id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    bio = EXCLUDED.bio,
    visibility = EXCLUDED.visibility,
    updated_at = now();

INSERT INTO buddy_assignments (id, student_id, buddy_id, active)
VALUES (
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000003',
    true
)
ON CONFLICT (id) DO UPDATE SET
    buddy_id = EXCLUDED.buddy_id,
    active = EXCLUDED.active,
    updated_at = now(),
    deleted_at = NULL;
