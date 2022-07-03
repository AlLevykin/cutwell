-- +goose Up
CREATE SCHEMA IF NOT EXISTS cutwell;
CREATE TABLE urls (
                      id varchar(9) PRIMARY KEY,
                      lnk text,
                      usr text
);
CREATE INDEX urls_users_idx ON urls ((lower(usr)));
-- +goose Down
DROP INDEX urls_users_idx;
DROP TABLE urls;