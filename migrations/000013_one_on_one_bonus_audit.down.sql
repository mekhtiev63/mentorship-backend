DROP INDEX IF EXISTS bonus_transactions_one_on_one_debit_uidx;
DROP INDEX IF EXISTS one_on_one_requests_bonus_reference_uidx;

ALTER TABLE one_on_one_requests
    DROP COLUMN IF EXISTS bonus_reference,
    DROP COLUMN IF EXISTS bonus_debited_at,
    DROP COLUMN IF EXISTS approved_at,
    DROP COLUMN IF EXISTS approved_by,
    DROP COLUMN IF EXISTS reject_reason;
