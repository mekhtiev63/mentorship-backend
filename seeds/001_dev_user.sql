-- Development seed user (password: changeme)
-- bcrypt hash for "changeme" (cost 12)

INSERT INTO users (id, email, password_hash, status)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin@example.com',
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
VALUES ('00000000-0000-0000-0000-000000000001', 'admin')
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role)
VALUES ('00000000-0000-0000-0000-000000000001', 'student')
ON CONFLICT DO NOTHING;

INSERT INTO profiles (user_id, display_name)
VALUES ('00000000-0000-0000-0000-000000000001', 'Admin')
ON CONFLICT (user_id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    updated_at = now();
