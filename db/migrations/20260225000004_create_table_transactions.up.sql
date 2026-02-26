CREATE TABLE transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id VARCHAR(100),
    account_id VARCHAR(100) NOT NULL REFERENCES accounts(account_id),
    merchant_id VARCHAR(100) NOT NULL REFERENCES merchants(merchant_id),
    amount DECIMAL(18,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_trace_id ON transactions(trace_id);
CREATE INDEX idx_transactions_account_id ON transactions(account_id);
CREATE INDEX idx_transactions_status ON transactions(status);
