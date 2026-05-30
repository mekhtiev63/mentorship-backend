# PostgreSQL schema v1.0

Baseline relational model for the Go mentorship platform. Migrations live in `migrations/` (golang-migrate).

## ER model

```mermaid
erDiagram
    users ||--o{ user_roles : has
    users ||--o| profiles : has
    users ||--o{ buddy_assignments : student
    users ||--o{ buddy_assignments : buddy
    users ||--o{ refresh_tokens : has
    users ||--o{ idempotency_keys : owns

    users ||--o{ student_block_progress : tracks
    roadmap_blocks ||--o{ materials : contains
    roadmap_blocks ||--o{ student_block_progress : for
    materials ||--o{ material_views : viewed

    users ||--o{ one_on_one_requests : student
    users ||--o{ one_on_one_requests : buddy
    calendar_events ||--o{ calendar_event_attendees : has
    users ||--o{ calendar_event_attendees : attends
    one_on_one_requests }o--o| calendar_events : scheduled

    users ||--o{ interviews : student
    users ||--o{ interviews : interviewer
    users ||--o{ final_assessments : student

    achievement_definitions ||--o{ user_achievements : granted
    users ||--o{ user_achievements : earns
    users ||--o{ activity_events : acts
    activity_events ||--o{ activity_feed_items : projects
    users ||--o{ activity_feed_items : receives

    users ||--o| bonus_accounts : has
    users ||--o{ bonus_transactions : logs

    users {
        uuid id PK
        citext email UK
        text password_hash
        user_status status
        timestamptz deleted_at
    }

    buddy_assignments {
        uuid id PK
        uuid student_id FK
        uuid buddy_id FK
        boolean active
        timestamptz deleted_at
    }

    roadmap_blocks {
        uuid id PK
        int sort_order
        roadmap_block_status status
        timestamptz deleted_at
    }

    materials {
        uuid id PK
        uuid block_id FK
        boolean required
        timestamptz deleted_at
    }

    student_block_progress {
        uuid student_id PK,FK
        uuid block_id PK,FK
        progress_status status
    }

    material_views {
        uuid student_id FK
        uuid material_id FK
    }

    bonus_accounts {
        uuid user_id PK,FK
        bigint balance
    }

    outbox_events {
        uuid id PK
        outbox_status status
    }
```

## Migration files

| Version | File | Content |
|---------|------|---------|
| 000001 | `000001_init` | `pgcrypto`, `citext` |
| 000002 | `000002_enums` | PostgreSQL ENUM types |
| 000003 | `000003_identity_user_profile` | users, roles, buddy, profiles, refresh, idempotency |
| 000004 | `000004_roadmap_progress` | roadmap, materials, progress, views |
| 000005 | `000005_sessions_calendar` | 1:1, calendar, interviews, final check |
| 000006 | `000006_gamification_activity` | achievements, activity, bonus, outbox |

## Business rules enforced in the database

| Invariant | Mechanism |
|-----------|-----------|
| Unique email | `users_email_active_uidx` (partial, `deleted_at IS NULL`) |
| One active buddy per student | `buddy_assignments_one_active_per_student_uidx` |
| Unique block sort order | `roadmap_blocks_sort_order_active_uidx` |
| Unique material order in block | `materials_block_sort_order_active_uidx` |
| One view per student/material | `material_views_student_material_uidx` |
| Non-negative bonus balance | `bonus_accounts_balance_non_negative` |
| Idempotent bonus convert | `bonus_transactions_idempotency_key_uidx` |
| One achievement grant per user/code | PK `(user_id, achievement_code)` |
| Idempotent achievement grant | `user_achievements_source_event_uidx` |
| One open final assessment per student | `final_assessments_one_open_per_student_uidx` |
| Roast after passed tech (baseline) | `final_assessments_roast_after_tech` CHECK |
| Publish/draft consistency | `roadmap_blocks_published_consistency` |
| Calendar time range | `calendar_events_time_range` |
| Idempotent API keys | `idempotency_keys_user_scope_key_uidx` |

Rules still enforced primarily in **application** layer (state machines, buddy scope, required views before submit, calendar overlap): transactions, row locks, and explicit checks.

## Index catalog

See [INDEXES.md](./INDEXES.md) for the full list and rationale.

## Soft delete

| Table | Column | Reason |
|-------|--------|--------|
| `users` | `deleted_at` | Retain referential history; login blocked via `status` + partial unique email |
| `buddy_assignments` | `deleted_at` | Historical assignments |
| `roadmap_blocks` | `deleted_at` | Catalog history with existing progress |
| `materials` | `deleted_at` | Preserve `material_views` |
| `calendar_events` | `deleted_at` | Cancel without losing audit trail |
| `achievement_definitions` | `deleted_at` | Retire codes without breaking grants |

## ISO/IEC 25010 mapping

- **Performance Efficiency:** partial indexes for hot lists (published roadmap, unread feed, pending outbox); composite indexes for buddy/student dashboards; no redundant wide indexes on JSONB except feed filters.
- **Reliability:** FK constraints, CHECK constraints, unique partial indexes for idempotency and single-active records; `outbox_events` for at-least-once side effects; `bonus_accounts.balance` guarded by CHECK (reconciled in TX with transactions).
- **Maintainability:** ENUM types in one migration; bounded-context-aligned migration files; documented ER and index purposes.
