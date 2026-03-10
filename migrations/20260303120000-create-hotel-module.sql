-- +migrate Up
CREATE TABLE hotels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    address VARCHAR(500) NOT NULL,
    city VARCHAR(255) NOT NULL,
    rating NUMERIC(3, 2) NOT NULL DEFAULT 0,
    check_in_time TIME NOT NULL,
    check_out_time TIME NOT NULL,
    default_child_max_age INT NOT NULL DEFAULT 11,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE hotel_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hotel_id UUID NOT NULL REFERENCES hotels(id) ON DELETE CASCADE,
    url VARCHAR(1000) NOT NULL,
    is_cover BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_hotel_images_hotel_id ON hotel_images(hotel_id);

CREATE TABLE hotel_payment_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hotel_id UUID NOT NULL REFERENCES hotels(id) ON DELETE CASCADE,
    payment_option VARCHAR(64) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE(hotel_id, payment_option)
);

CREATE INDEX idx_hotel_payment_options_hotel_id ON hotel_payment_options(hotel_id);

-- +migrate Down
DROP TABLE IF EXISTS hotel_payment_options;
DROP TABLE IF EXISTS hotel_images;
DROP TABLE IF EXISTS hotels;
