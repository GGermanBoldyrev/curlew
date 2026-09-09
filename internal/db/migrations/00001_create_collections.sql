-- +goose Up
CREATE TABLE collections (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER REFERENCES collections(id) ON DELETE CASCADE,
    name      TEXT NOT NULL
);

CREATE UNIQUE INDEX collections_sibling_name
    ON collections(COALESCE(parent_id, 0), name COLLATE NOCASE);

-- +goose Down
DROP INDEX collections_sibling_name;
DROP TABLE collections;
