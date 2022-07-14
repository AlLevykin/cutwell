-- +goose Up
DELETE FROM urls a
WHERE a.id <> (SELECT min(b.id)
                 FROM   urls b
                 WHERE  a.lnk = b.lnk);
CREATE UNIQUE INDEX urls_links_idx ON urls ((lower(lnk)));
-- +goose Down
DROP INDEX urls_links_idx;
