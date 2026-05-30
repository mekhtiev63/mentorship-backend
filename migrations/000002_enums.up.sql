-- Domain enumerations as PostgreSQL types (explicit, index-friendly CHECK boundaries).

CREATE TYPE user_status AS ENUM ('active', 'disabled');

CREATE TYPE app_role AS ENUM ('student', 'buddy', 'admin');

CREATE TYPE profile_visibility AS ENUM ('public', 'buddies_only', 'private');

CREATE TYPE roadmap_block_status AS ENUM ('draft', 'published');

CREATE TYPE material_type AS ENUM ('video', 'article', 'task', 'link');

CREATE TYPE progress_status AS ENUM (
    'not_started',
    'in_progress',
    'awaiting_approval',
    'approved',
    'rejected'
);

CREATE TYPE one_on_one_status AS ENUM (
    'pending',
    'accepted',
    'scheduled',
    'completed',
    'cancelled'
);

CREATE TYPE interview_kind AS ENUM ('mock', 'real');

CREATE TYPE interview_status AS ENUM (
    'scheduled',
    'in_progress',
    'completed',
    'cancelled'
);

CREATE TYPE interview_outcome AS ENUM ('pending', 'passed', 'failed');

CREATE TYPE assessment_status AS ENUM (
    'not_started',
    'in_progress',
    'passed',
    'failed',
    'cancelled'
);

CREATE TYPE calendar_related_type AS ENUM ('one_on_one', 'interview', 'other');

CREATE TYPE bonus_transaction_type AS ENUM ('credit', 'debit', 'convert');

CREATE TYPE outbox_status AS ENUM ('pending', 'processing', 'done', 'failed');
