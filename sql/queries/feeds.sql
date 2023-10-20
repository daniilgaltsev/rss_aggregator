-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeed :one
SELECT * FROM feeds WHERE id = $1;

-- name: GetNextFeedsToFetch :many
SELECT * FROM feeds
WHERE (
    (last_fetched_at < NOW()) OR
    (last_fetched_at IS NULL)
)
ORDER BY last_fetched_at NULLS FIRST
LIMIT $1;

-- name: UpdateLastFetchedAt :exec
UPDATE feeds
SET
 last_fetched_at = NOW(),
 updated_at = NOW()
WHERE id=$1;
