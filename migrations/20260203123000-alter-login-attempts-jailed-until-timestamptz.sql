-- +migrate Up
ALTER TABLE login_attempts
    ALTER COLUMN jailed_until TYPE TIMESTAMPTZ
    USING jailed_until AT TIME ZONE 'UTC';

-- +migrate Down
ALTER TABLE login_attempts
    ALTER COLUMN jailed_until TYPE TIMESTAMP
    USING jailed_until AT TIME ZONE 'UTC';
