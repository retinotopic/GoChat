package db

import (
	"context"
	"errors"
	"fmt"
	"log"
)

// method for safely creating unique duo room
func (c *PostgresClient) CreateDuoRoom(flowjson *FlowJSON) {
	c.IsDuoRoomExist(flowjson)
	if flowjson.Err != nil {
		c.CreateRoom(flowjson)
		_, flowjson.Err = flowjson.Tx.Exec(context.Background(), `INSERT INTO duo_rooms (user_id1, user_id2,room_id)`, flowjson.Users[0], flowjson.Users[1], flowjson.Rooms[0])
		c.AddUsersToRoom(flowjson)
	} else {
		c.AddUsersToRoom(flowjson)
		return
	}

}
func (c *PostgresClient) IsDuoRoomExist(flowjson *FlowJSON) {
	var row = flowjson.Tx.QueryRow(context.Background(), `SELECT room_id
		FROM duo_rooms
		WHERE user_id1 = $1 AND user_id2 = $2;`, flowjson.Users[0], flowjson.Users[1])
	flowjson.Err = row.Scan(&flowjson.Rooms[0])
}
func (c *PostgresClient) CreateRoom(flowjson *FlowJSON) {
	var roomID uint32
	var is_group bool
	if flowjson.Mode != "createDuoRoom" {
		is_group = true
	}
	// create new room and return room id
	flowjson.Err = flowjson.Tx.QueryRow(context.Background(), "INSERT INTO rooms (name,is_group) VALUES ($1,$2) RETURNING room_id", flowjson.Name, is_group).Scan(&roomID)
	if flowjson.Err != nil {
		log.Println("Error inserting room:", flowjson.Err)
		return
	}
	flowjson.Rooms[0] = roomID
	c.AddUsersToRoom(flowjson)
}
func (c *PostgresClient) AddUsersToRoom(flowjson *FlowJSON) {
	if flowjson.Mode == "AddUsersToRoom" {
		if err := c.Conn.QueryRow(context.Background(), `SELECT 1 FROM rooms WHERE room_id = $1 AND created_by_user_id = $2`, flowjson.Rooms[0], flowjson.Users[0]).Scan(new(int)); err != nil {
			flowjson.Err = errors.New("you have no permission to add users to this room")
			return
		}
	}
	query := `INSERT INTO room_users_info (user_id, room_id)
			SELECT users_to_add.user_id, $1
			FROM (SELECT unnest($2) AS user_id) AS users_to_add
			JOIN users u ON u.user_id = users_to_add.user_id AND %s
			JOIN rooms r ON r.room_id = $1 AND r.is_group = $3
			LEFT JOIN blocked_users bu ON bu.blocked_by_user_id = $4 AND bu.blocked_user_id = users_to_add.user_id
			WHERE bu.blocked_by_user_id IS NULL;`
	var condition string
	var is_group bool
	if flowjson.Mode == "createDuoRoom" {
		condition = "u.allow_direct_messages = true"
	} else {
		condition = "u.allow_group_invites = true"
		is_group = true
	}

	query = fmt.Sprintf(query, condition)

	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), query, flowjson.Rooms[0], flowjson.Users, is_group, c.UserID)

}
func (c *PostgresClient) DeleteUsersFromRoom(flowjson *FlowJSON) {
	if flowjson.Mode == "DeleteUsersFromRoom" {
		if len(flowjson.Users) != 1 && flowjson.Users[0] != c.UserID {
			if err := c.Conn.QueryRow(context.Background(), `SELECT 1 FROM rooms WHERE room_id = $1 AND created_by_user_id = $2`, flowjson.Rooms[0], flowjson.Users[0]).Scan(new(int)); err != nil {
				flowjson.Err = errors.New("you have no permission to delete users from this room")
				return
			}
		}
	}
	query := `DELETE FROM room_users_info
	WHERE user_id = ANY($1);
	AND room_id IN (
		SELECT room_id
		FROM rooms 
		WHERE room_id = $2 AND is_group = $3
	);`
	var is_group bool

	if flowjson.Mode != "BlockUser" {
		is_group = true
	}
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), query, flowjson.Users, flowjson.Rooms[0], is_group)
}
