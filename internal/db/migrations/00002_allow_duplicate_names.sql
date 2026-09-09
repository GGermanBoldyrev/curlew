-- +goose Up
DROP INDEX collections_sibling_name;

-- +goose Down
CREATE UNIQUE INDEX collections_sibling_name
    ON collections(COALESCE(parent_id, 0), name COLLATE NOCASE);
