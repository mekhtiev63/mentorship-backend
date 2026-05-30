-- Progress: require reject reason when status is rejected.

ALTER TABLE student_block_progress
    ADD CONSTRAINT student_block_progress_reject_reason_when_rejected CHECK (
        status <> 'rejected'
        OR (reject_reason IS NOT NULL AND length(trim(reject_reason)) > 0)
    );
