package db

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

// method for safely creating unique duo room
func (p *PgClient) CreateDuoRoom(ctx context.Context, flowjson *FlowJSON) {
	p.IsDuoRoomExist(ctx, flowjson)
	if flowjson.Err != nil {
		log.Println("IsDuoRoomExist error")
		return
	}
	if flowjson.Room == 0 {
		p.CreateRoom(ctx, flowjson)
		if flowjson.Err != nil {
			log.Println("CreateRoom error")
			return
		}
		_, flowjson.Err = flowjson.Tx.Exec(context.Background(), `INSERT INTO duo_rooms (user_id1,user_id2,room_id) VALUES ($1,$2,$3)`, flowjson.Users[0], flowjson.Users[1], flowjson.Room)
	} else {
		p.AddUsersToRoom(ctx, flowjson)
	}
}
func (p *PgClient) IsDuoRoomExist(ctx context.Context, flowjson *FlowJSON) {
	var rows pgx.Rows
	rows, flowjson.Err = flowjson.Tx.Query(context.Background(), `SELECT room_id
		FROM duo_rooms
		WHERE (user_id1 = $1 AND user_id2 = $2) OR (user_id2 = $1 AND user_id1 = $2) ;`, flowjson.Users[0], flowjson.Users[1])
	if flowjson.Err != nil {
		return
	}
	if rows.Next() {
		flowjson.Err = rows.Scan(&flowjson.Room)
		rows.Close()
	}
}
func (p *PgClient) CreateRoom(ctx context.Context, flowjson *FlowJSON) {
	var is_group bool
	if flowjson.Mode != "CreateDuoRoom" {
		is_group = true
	}
	// create new room and return room id
	flowjson.Err = flowjson.Tx.QueryRow(context.Background(), "INSERT INTO rooms (name,is_group,created_by_user_id) VALUES ($1,$2,$3) RETURNING room_id", flowjson.Name, is_group, p.UserID).Scan(&flowjson.Room)
	if flowjson.Err != nil {
		log.Println("Error inserting room:", flowjson.Err)
		return
	}
	p.AddUsersToRoom(ctx, flowjson)
}
func (p *PgClient) AddUsersToRoom(ctx context.Context, flowjson *FlowJSON) {
	var rows pgx.Rows
	if flowjson.Mode == "AddUsersToRoom" {
		if err := flowjson.Tx.QueryRow(context.Background(), `SELECT 1 FROM rooms WHERE room_id = $1 AND created_by_user_id = $2`, flowjson.Room, p.UserID).Scan(new(int)); err != nil {
			flowjson.Err = errors.New("you have no permission to add users to this room")
			log.Println("you have no permission to add users to this room")
			return
		}
	}
	query := `INSERT INTO room_users_info (room_id,user_id)
			SELECT $1,users_to_add.user_id
			FROM (SELECT unnest($2::int[]) AS user_id) AS users_to_add
			JOIN users u ON u.user_id = users_to_add.user_id AND %s
			JOIN rooms r ON r.room_id = $1 AND r.is_group = $3
			LEFT JOIN blocked_users bu ON (bu.blocked_by_user_id = users_to_add.user_id AND bu.blocked_user_id = $4) 
			OR (bu.blocked_by_user_id = $4 AND bu.blocked_user_id = users_to_add.user_id )
			WHERE bu.blocked_by_user_id IS NULL RETURNING user_id;`
	var condition string
	var is_group bool
	if flowjson.Mode == "CreateDuoRoom" {
		condition = "u.allow_direct_messages = true"
	} else {
		condition = "u.allow_group_invites = true"
		is_group = true
	}
	query = fmt.Sprintf(query, condition)
	rows, flowjson.Err = flowjson.Tx.Query(context.Background(), query, flowjson.Room, flowjson.Users, is_group, p.UserID)
	counter := 0
	for rows.Next() {
		counter++
	}
	if len(flowjson.Users) != counter {
		flowjson.Err = errors.New("at least one user cannot be added to the room")
	}
}
func (p *PgClient) DeleteUsersFromRoom(ctx context.Context, flowjson *FlowJSON) {
	if flowjson.Mode == "DeleteUsersFromRoom" {
		if len(flowjson.Users) != 1 || flowjson.Users[0] != p.UserID {
			if err := flowjson.Tx.QueryRow(context.Background(), `SELECT 1 FROM rooms WHERE room_id = $1 AND created_by_user_id = $2`, flowjson.Room, p.UserID).Scan(new(int)); err != nil {
				flowjson.Err = errors.New("you have no permission to delete users from this room")
				return
			}
		}
	}
	query := `DELETE FROM room_users_info
	WHERE user_id = ANY($1)
	AND room_id IN (
		SELECT room_id
		FROM rooms 
		WHERE room_id = $2 AND is_group = $3
	);`
	var is_group bool

	if flowjson.Mode != "BlockUser" {
		is_group = true
	}
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), query, flowjson.Users, flowjson.Room, is_group)
}
