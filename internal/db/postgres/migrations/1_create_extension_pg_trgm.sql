-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


-- +goose Down
DROP EXTENSION IF EXISTS pg_trgm;