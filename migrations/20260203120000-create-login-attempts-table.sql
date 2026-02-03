-- +migrate Up
CREATE TABLE login_attempts (
    email VARCHAR(255) PRIMARY KEY,
    failed_count INTEGER NOT NULL DEFAULT 0,
    jailed_until TIMESTAMP NULL
);

-- +migrate Down
DROP TABLE IF EXISTS login_attempts;
