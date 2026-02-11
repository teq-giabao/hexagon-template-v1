-- +migrate Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_movies_title_trgm ON movies USING GIN (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_movies_genres_trgm ON movies USING GIN (genres gin_trgm_ops);

-- +migrate Down
DROP INDEX IF EXISTS idx_movies_genres_trgm;
DROP INDEX IF EXISTS idx_movies_title_trgm;
-- NOTE: do not drop extension in down migration because it may be used elsewhere.
