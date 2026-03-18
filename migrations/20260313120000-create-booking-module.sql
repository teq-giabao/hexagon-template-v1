-- +migrate Up
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hotel_id UUID NOT NULL REFERENCES hotels(id) ON DELETE CASCADE,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    nights INT NOT NULL,
    room_count INT NOT NULL,
    guest_count INT NOT NULL,
    nightly_price NUMERIC(12, 2) NOT NULL,
    total_price NUMERIC(12, 2) NOT NULL,
    status VARCHAR(32) NOT NULL,
    payment_option VARCHAR(64),
    payment_status VARCHAR(32) NOT NULL DEFAULT 'unpaid',
    hold_expires_at TIMESTAMPTZ,
    payment_deadline TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    cancellation_fee NUMERIC(12, 2) NOT NULL DEFAULT 0,
    refund_amount NUMERIC(12, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bookings_hotel_id ON bookings(hotel_id);
CREATE INDEX idx_bookings_room_id ON bookings(room_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_hold_expires_at ON bookings(hold_expires_at);
CREATE INDEX idx_bookings_payment_deadline ON bookings(payment_deadline);

-- +migrate Down
DROP TABLE IF EXISTS bookings;
