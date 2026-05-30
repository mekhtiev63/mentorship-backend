ALTER TABLE one_on_one_requests
    ADD COLUMN IF NOT EXISTS reject_reason TEXT,
    ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS bonus_debited_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS bonus_reference TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS one_on_one_requests_bonus_reference_uidx
    ON one_on_one_requests (bonus_reference)
    WHERE bonus_reference IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS bonus_transactions_one_on_one_debit_uidx
    ON bonus_transactions (user_id, reference)
    WHERE type = 'debit' AND reference LIKE 'one_on_one:%';
