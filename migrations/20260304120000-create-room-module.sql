-- +migrate Up
CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hotel_id UUID NOT NULL REFERENCES hotels(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_price NUMERIC(12, 2) NOT NULL,
    max_adult INT NOT NULL,
    max_child INT NOT NULL DEFAULT 0,
    max_occupancy INT NOT NULL,
    bed_options JSONB NOT NULL DEFAULT '[]'::jsonb,
    size_sqm INT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rooms_hotel_id ON rooms(hotel_id);

CREATE TABLE room_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    url VARCHAR(1000) NOT NULL,
    is_cover BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_room_images_room_id ON room_images(room_id);

CREATE TABLE room_inventories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    total_inventory INT NOT NULL,
    held_inventory INT NOT NULL DEFAULT 0,
    booked_inventory INT NOT NULL DEFAULT 0,
    UNIQUE(room_id, date)
);

CREATE INDEX idx_room_inventories_room_id ON room_inventories(room_id);

CREATE TABLE room_amenities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE room_amenity_maps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    amenity_id UUID NOT NULL REFERENCES room_amenities(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(room_id, amenity_id)
);

CREATE INDEX idx_room_amenity_maps_room_id ON room_amenity_maps(room_id);
CREATE INDEX idx_room_amenity_maps_amenity_id ON room_amenity_maps(amenity_id);

-- +migrate Down
DROP TABLE IF EXISTS room_amenity_maps;
DROP TABLE IF EXISTS room_amenities;
DROP TABLE IF EXISTS room_inventories;
DROP TABLE IF EXISTS room_images;
DROP TABLE IF EXISTS rooms;
