-- +goose Up
CREATE SCHEMA IF NOT EXISTS cutwell;
CREATE TABLE urls (
                      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                      url text,
                      shorten text,
                      "user" text
);
CREATE INDEX urls_users_idx ON urls ((lower("user")));
-- +goose Down
DROP INDEX urls_users_idx;
DROP TABLE urls;