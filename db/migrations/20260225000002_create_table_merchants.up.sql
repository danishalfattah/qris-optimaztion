CREATE TABLE merchants (
    merchant_id VARCHAR(100) PRIMARY KEY,
    merchant_name VARCHAR(255) NOT NULL,
    mcc VARCHAR(10) NOT NULL,
    city VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Seed sample merchants for testing
INSERT INTO merchants (merchant_id, merchant_name, mcc, city) VALUES
('MICH-001', 'Toko Berkah Mandiri', '2741', 'Malang'),
('MICH-002', 'M Ivan Store', '2741', 'Jakarta Timur');
