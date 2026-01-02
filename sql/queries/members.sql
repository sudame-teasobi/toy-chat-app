
-- name: GetMember :one
SELECT * FROM chat_room_members WHERE id = ? LIMIT 1;

-- name: CreateMember :execresult
INSERT INTO chat_room_members (id, user_id, chat_room_id) VALUES (?, ?, ?);

