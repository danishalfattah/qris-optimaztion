CREATE TABLE accounts (
    account_id VARCHAR(100) PRIMARY KEY,
    balance DECIMAL(18,2) NOT NULL DEFAULT 0,
    currency VARCHAR(5) NOT NULL DEFAULT 'IDR',
    pin_hash VARCHAR(255) NOT NULL,
    version INT NOT NULL DEFAULT 0
);

-- Seed sample account for testing (pin: 123456, bcrypt hash)
INSERT INTO accounts (account_id, balance, currency, pin_hash, version) VALUES
('user_123', 999999999.00, 'IDR', '$2a$10$LuiFuShDopsnneELl665GuwPgGVggEgILH7nq.rPlM7CyydZJhiwS', 0);
