CREATE TABLE api_clients (
    client_id VARCHAR(100) PRIMARY KEY,
    client_secret VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed default API client for testing
INSERT INTO api_clients (client_id, client_secret, status) VALUES
('MK-9921-X', 'super-secret-key-123', 'ACTIVE');
