-- +migrate Up
CREATE TABLE IF NOT EXISTS movies (
    id BIGSERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL UNIQUE,
    title TEXT NOT NULL,
    genres TEXT NOT NULL DEFAULT '',
    search_vector tsvector GENERATED ALWAYS AS (
        to_tsvector('english', coalesce(title, '') || ' ' || coalesce(genres, ''))
    ) STORED
);

CREATE INDEX IF NOT EXISTS idx_movies_search_vector ON movies USING GIN (search_vector);

-- +migrate Down
DROP INDEX IF EXISTS idx_movies_search_vector;
DROP TABLE IF EXISTS movies;
