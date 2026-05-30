-- Idempotent achievement bonus credits per user and reference.

CREATE UNIQUE INDEX IF NOT EXISTS bonus_transactions_achievement_credit_uidx
    ON bonus_transactions (user_id, reference)
    WHERE type = 'credit' AND reference LIKE 'achievement:%';
