-- name: GetMessage :one
SELECT * FROM messages WHERE id = ? LIMIT 1;

-- name: PostMessage :execresult
INSERT INTO messages (id, user_id, chat_room_id, body) VALUES (?, ?, ?, ?);

