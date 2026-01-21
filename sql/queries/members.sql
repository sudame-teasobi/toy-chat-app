
-- name: GetMember :one
SELECT * FROM chat_room_members WHERE id = ? LIMIT 1;

-- name: GetMembersByChatRoomID :many
SELECT * FROM chat_room_members WHERE chat_room_id = ?;

-- name: CreateMember :execresult
INSERT INTO chat_room_members (id, user_id, chat_room_id) VALUES (?, ?, ?);

-- name: GetMemberByChatRoomAndUser :one
SELECT * FROM chat_room_members WHERE chat_room_id = ? AND user_id = ? LIMIT 1;
