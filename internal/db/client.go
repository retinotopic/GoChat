package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func (c *PostgresClient) GetAllRooms(ctx context.Context, flowjson *FlowJSON) {
	defer c.Mutex.Unlock()
	var Rows pgx.Rows
	Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT r.room_id
		FROM room_users_info ru JOIN rooms r ON ru.room_id = r.room_id
		WHERE ru.user_id = $1 
		ORDER BY r.last_activity DESC;
		`, c.UserID)
	err := Rows.Err()
	if err != nil {
		log.Println("query timeout", flowjson.Err)
		flowjson.Err = err
		return
	}
	c.Mutex.Lock()
	for Rows.Next() {
		err := Rows.Scan(&flowjson.Room)
		if err != nil {
			log.Println("Error scanning rows:", err)
			flowjson.Err = err
			return
		}
		c.Chan <- *flowjson
		c.RoomsPagination = append(c.RoomsPagination, flowjson.Room)
	}
}

// load messages from a room
func (c *PostgresClient) GetMessagesFromRoom(ctx context.Context, flowjson *FlowJSON) {
	var Rows pgx.Rows
	Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT payload,user_id,
		FROM messages 
		WHERE room_id = $1 AND message_id < $2
		ORDER BY message_id DESC`, flowjson.Room, flowjson.Message_id)
	if flowjson.Err != nil {
		log.Println("Error getting messages from this rooms:", flowjson.Err)
		return
	}
	c.toChannel(flowjson, Rows, flowjson.Message, flowjson.Users[0])
}

func (c *PostgresClient) GetNextRooms(ctx context.Context, flowjson *FlowJSON) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	var arrayrooms []uint32
	for i := c.PaginationOffset; i < c.PaginationOffset+30; i++ {
		arrayrooms = append(arrayrooms, c.RoomsPagination[i])
	}
	var Rows pgx.Rows
	Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT room_id,name FROM rooms WHERE room_id IN ($1)`, arrayrooms)
	if flowjson.Err != nil {
		log.Println("Error getting messages from this rooms:", flowjson.Err)
		return
	}
	c.toChannel(flowjson, Rows, flowjson.Room, flowjson.Name)
	if flowjson.Err == nil {
		c.PaginationOffset += 30
	}
}

func (c *PostgresClient) GetRoomUsersInfo(ctx context.Context, flowjson *FlowJSON) {
	var Rows pgx.Rows
	Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT u.user_id,u.name
		FROM users u JOIN room_users_info ru ON ru.user_id = u.user_id
		WHERE ru.room_id = $1`, c.UserID)
	if flowjson.Err != nil {
		log.Println("Error getting room info", flowjson.Err)
		return
	}
	c.toChannel(flowjson, Rows, &flowjson.Users[0], &flowjson.Name)
}

func (c *PostgresClient) FindUsers(ctx context.Context, flowjson *FlowJSON) {
	var Rows pgx.Rows
	Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT user_id,username FROM users WHERE username ILIKE $1 LIMIT 20`, flowjson.Name+"%")
	c.toChannel(flowjson, Rows, &flowjson.Users[0], &flowjson.Name)
}
func (c *PostgresClient) toChannel(flowjson *FlowJSON, rows pgx.Rows, dest ...any) {
	err := rows.Err() // checking for query timeout
	if err != nil {
		flowjson.Err = err
		return
	}
	for rows.Next() {
		err := rows.Scan(dest...)
		if err != nil {
			log.Println("Error scanning rows:", err)
			flowjson.Err = err
			return
		}
		c.Chan <- *flowjson
	}
}
