-- name: GetMessage :one
SELECT * FROM messages WHERE id = ? LIMIT 1;

-- name: ListMessages :many
SELECT * FROM messages ORDER BY created_at DESC LIMIT ?;

-- name: ListMessagesByUser :many
SELECT * FROM messages WHERE user_id = ? ORDER BY created_at DESC;

-- name: CreateMessage :execresult
INSERT INTO messages (user_id, content) VALUES (?, ?);

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = ?;
