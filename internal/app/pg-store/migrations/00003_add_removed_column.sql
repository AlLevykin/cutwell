-- +goose Up
ALTER TABLE urls ADD COLUMN removed BOOLEAN DEFAULT FALSE;
-- +goose Down
ALTER TABLE urls DROP COLUMN removed;