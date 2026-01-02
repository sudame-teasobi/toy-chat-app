
-- name: GetChatRoom :one
SELECT * FROM chat_rooms WHERE id = ? LIMIT 1;

-- name: CreateChatRoom :execresult
INSERT INTO chat_rooms (id, name) VALUES (?, ?);

