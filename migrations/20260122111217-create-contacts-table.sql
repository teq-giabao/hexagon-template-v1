
-- +migrate Up
CREATE TABLE contacts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(255) NOT NULL
);

-- +migrate Down
DROP TABLE contacts;