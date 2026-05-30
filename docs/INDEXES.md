# Index catalog and rationale

## users

| Index | Columns | Type | Rationale |
|-------|---------|------|-----------|
| `users_email_active_uidx` | `email` | UNIQUE partial (`deleted_at IS NULL`) | Login lookup by email; allows re-registration after soft delete without breaking history |
| `users_status_idx` | `status` | partial active users | Admin filters, health checks on active population |

## user_roles

| Index | Columns | Rationale |
|-------|---------|-----------|
| `user_roles_role_user_idx` | `(role, user_id)` | RBAC queries: list buddies/admins without scanning all users |

## buddy_assignments

| Index | Columns | Rationale |
|-------|---------|-----------|
| `buddy_assignments_one_active_per_student_uidx` | `student_id` | UNIQUE partial (`active`, not deleted) — **invariant #7** one active buddy |
| `buddy_assignments_buddy_active_idx` | `buddy_id` | partial active | Buddy dashboard: assigned students |
| `buddy_assignments_student_history_idx` | `(student_id, created_at DESC)` | Assignment history audit |

## refresh_tokens

| Index | Columns | Rationale |
|-------|---------|-----------|
| `refresh_tokens_token_hash_uidx` | `token_hash` | O(1) refresh validation |
| `refresh_tokens_user_active_idx` | `(user_id, expires_at DESC)` | partial non-revoked | Revoke-all / list active sessions per user |

## idempotency_keys

| Index | Columns | Rationale |
|-------|---------|-----------|
| `idempotency_keys_user_scope_key_uidx` | `(user_id, scope, idempotency_key)` | **Invariant #56** duplicate POST protection |
| `idempotency_keys_expires_at_idx` | `expires_at` | TTL cleanup job |

## roadmap_blocks

| Index | Columns | Rationale |
|-------|---------|-----------|
| `roadmap_blocks_sort_order_active_uidx` | `sort_order` | UNIQUE partial — **invariant #16** order among active blocks |
| `roadmap_blocks_published_list_idx` | `sort_order` | partial published | Student catalog `GET /roadmap/blocks` — index-only friendly ordering |

## materials

| Index | Columns | Rationale |
|-------|---------|-----------|
| `materials_block_sort_order_active_uidx` | `(block_id, sort_order)` | UNIQUE partial | Stable ordering within block |
| `materials_block_active_idx` | `block_id` | partial | Load block with materials (avoid N+1 via `WHERE block_id = $1`) |

## student_block_progress

| Index | Columns | Rationale |
|-------|---------|-----------|
| PK | `(student_id, block_id)` | Natural key; efficient upsert on progress |
| `student_block_progress_student_status_idx` | `(student_id, status)` | Student dashboard by status |
| `student_block_progress_block_awaiting_idx` | `(block_id, submitted_at)` | partial `awaiting_approval` | Buddy approval queue |

## material_views

| Index | Columns | Rationale |
|-------|---------|-----------|
| `material_views_student_material_uidx` | `(student_id, material_id)` | UNIQUE — **invariant #20** idempotent view |
| `material_views_idempotency_uidx` | `(student_id, idempotency_key)` | UNIQUE partial | Header-based idempotency replay |
| `material_views_material_idx` | `material_id` | Aggregate “who viewed” / analytics |

## one_on_one_requests

| Index | Columns | Rationale |
|-------|---------|-----------|
| `one_on_one_requests_student_status_idx` | `(student_id, status, created_at DESC)` | Student list |
| `one_on_one_requests_buddy_status_idx` | `(buddy_id, status, created_at DESC)` | Buddy inbox |

## calendar_events / attendees

| Index | Columns | Rationale |
|-------|---------|-----------|
| `calendar_events_organizer_range_idx` | `(organizer_id, starts_at)` | partial | Range queries; overlap detection scoped to organizer |
| `calendar_events_related_idx` | `(related_type, related_id)` | Link 1:1 / interview to event |
| `calendar_event_attendees_user_idx` | `user_id` | Attendee visibility — **invariant #35** |

Overlap prevention uses `organizer_id` + time range in a transaction (application); index supports `WHERE organizer_id = $1 AND starts_at < $3 AND ends_at > $2`.

## interviews

| Index | Columns | Rationale |
|-------|---------|-----------|
| `interviews_student_status_idx` | `(student_id, status, scheduled_at DESC)` | Student scope |
| `interviews_interviewer_status_idx` | `(interviewer_id, status, scheduled_at DESC)` | Buddy/admin workload |

## final_assessments

| Index | Columns | Rationale |
|-------|---------|-----------|
| `final_assessments_one_open_per_student_uidx` | `student_id` | UNIQUE partial (`cancelled_at IS NULL`) — **invariant #39** |
| `final_assessments_tech_reviewer_idx` | `tech_reviewer_id` | Reviewer queue |

## achievements / activity

| Index | Columns | Rationale |
|-------|---------|-----------|
| `achievement_definitions_active_idx` | `code` | partial | Catalog without deleted codes |
| `user_achievements_source_event_uidx` | `(source_event_id, achievement_code)` | UNIQUE — **invariant #44** |
| `user_achievements_user_granted_idx` | `(user_id, granted_at DESC)` | Profile / me achievements |
| `activity_events_created_at_idx` | `created_at DESC` | Admin/audit timelines |
| `activity_events_actor_created_idx` | `(actor_id, created_at DESC)` | Actor history |
| `activity_feed_items_event_user_uidx` | `(event_id, user_id)` | UNIQUE — **invariant #47** no duplicate projection |
| `activity_feed_items_user_unread_idx` | `(user_id, created_at DESC)` | partial unread | Notification bell |

## bonus

| Index | Columns | Rationale |
|-------|---------|-----------|
| `bonus_transactions_idempotency_key_uidx` | `idempotency_key` | UNIQUE partial — **invariant #51** |
| `bonus_transactions_user_created_idx` | `(user_id, created_at DESC)` | Paginated ledger |

## outbox

| Index | Columns | Rationale |
|-------|---------|-----------|
| `outbox_events_pending_idx` | `created_at` | partial pending/failed | Worker poll without full table scan |

## Primary / foreign keys

All relationships use explicit `REFERENCES` with `ON DELETE` chosen per aggregate:

- **CASCADE:** child rows owned by user (`profiles`, `user_roles`, feed items).
- **RESTRICT:** business facts that must not disappear (`progress`, `views`, transactions).
- **SET NULL:** optional reviewers / calendar link on cancel.

This supports **Reliability** (no orphan progress) and **Maintainability** (predictable delete semantics).
