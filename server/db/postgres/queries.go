package db

import "fmt"

var allowGroupInvites = "u.allow_group_invites = true"
var allowDirectMessages = "u.allow_direct_messages = true"

var addUsersToRoom = `
WITH addusers AS 
(INSERT INTO room_users_info (room_id,user_id)
SELECT $1,users_to_add.user_id
FROM (SELECT unnest($2::int[]) AS user_id) AS users_to_add
JOIN users u ON u.user_id = users_to_add.user_id AND %s
JOIN rooms r ON r.room_id = $1 AND r.is_group = $3
LEFT JOIN blocked_users bu ON (bu.blocked_by_user_id = users_to_add.user_id AND bu.blocked_user_id = $4) 
OR (bu.blocked_by_user_id = $4 AND bu.blocked_user_id = users_to_add.user_id )
WHERE bu.blocked_by_user_id IS NULL RETURNING user_id,room_id)

SELECT ad.user_id,u.user_name,ad.room_id,r.room_name,r.is_group,r.created_by_user_id FROM addusers ad JOIN users u ON ad.user_id = u.user_id JOIN rooms r ON ad.room_id = r.room_id`

// there can be only two prepared cached statements for addUsersToRoom: allowGroupInvites and allowDirectMessages
var addUsersToRoomGroup = fmt.Sprintf(addUsersToRoom, allowGroupInvites)
var addUsersToRoomDirect = fmt.Sprintf(addUsersToRoom, allowDirectMessages)

var deleteUsersFromRoom = `
WITH delusers AS
(DELETE FROM room_users_info
WHERE user_id = ANY($1)
AND room_id IN (
	SELECT room_id
	FROM rooms 
	WHERE room_id = $2 AND is_group = $3
) RETURNING user_id,room_id ) 
 
SELECT de.user_id,u.user_name,de.room_id,r.room_name,r.is_group FROM delusers de JOIN users u ON de.user_id = u.user_id JOIN rooms r ON de.room_id = r.room_id`

var CreateRoom = "INSERT INTO rooms (room_name,is_group,created_by_user_id) VALUES ($1,$2,$3) RETURNING room_id"

var OldMessages = "m.message_id < $2"
var NewMessages = "m.message_id > $2"

var getMessagesFromRoom = `SELECT m.message_payload,m.user_id,u.user_name,m.message_id,m.room_id
FROM messages m JOIN users u ON u.user_id = m.user_id WHERE m.room_id = $1 AND %s ORDER BY m.message_id DESC LIMIT 20`

var getOldMessages = fmt.Sprintf(getMessagesFromRoom, OldMessages)
var getNewMessages = fmt.Sprintf(getMessagesFromRoom, NewMessages)
