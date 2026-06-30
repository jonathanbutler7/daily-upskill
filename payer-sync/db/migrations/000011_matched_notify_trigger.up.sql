-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION notify_reconciled_payment_matched()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'MATCHED' AND (OLD.status IS NULL OR OLD.status <> 'MATCHED') THEN
        PERFORM pg_notify('reconciled_payment_matched', NEW.reconciled_payment_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_reconciled_payment_matched
AFTER INSERT OR UPDATE OF status ON reconciled_payments
FOR EACH ROW EXECUTE FUNCTION notify_reconciled_payment_matched();

-- +goose Down

DROP TRIGGER IF EXISTS trg_reconciled_payment_matched ON reconciled_payments;
DROP FUNCTION IF EXISTS notify_reconciled_payment_matched();
