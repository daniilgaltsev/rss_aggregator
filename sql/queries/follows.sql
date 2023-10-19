-- name: CreateFollow :one
INSERT INTO follows (id, created_at, updated_at, feed_id, user_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFollow :one
SELECT * FROM follows WHERE id = $1;

-- name: DeleteFollow :exec
DELETE FROM follows WHERE id = $1;
