package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func (c *PostgresClient) GetAllRooms(flowjson *FlowJSON) {
	var Rows pgx.Rows
	Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT r.room_id
		FROM room_users_info ru JOIN rooms r ON ru.room_id = r.room_id
		WHERE ru.user_id = $1 AND is_visible = true 
		ORDER BY r.last_activity DESC;
		`, c.UserID)
	err := Rows.Err()
	if err != nil {
		log.Println("query timeout", flowjson.Err)
		flowjson.Err = err
		return
	}
	for Rows.Next() {
		err := Rows.Scan(&flowjson.Rooms[0])
		if err != nil {
			log.Println("Error scanning rows:", err)
			flowjson.Err = err
			return
		}
		c.Chan <- *flowjson
		c.RoomsPagination = append(c.RoomsPagination, flowjson.Rooms[0])
	}
}

// load messages from a room
func (c *PostgresClient) GetMessagesFromRoom(flowjson *FlowJSON) {
	var Rows pgx.Rows
	Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT payload,user_id,
		FROM messages 
		WHERE room_id = $1 AND message_id < $2
		ORDER BY message_id DESC`, flowjson.Rooms[0], flowjson.Message_id)
	if flowjson.Err != nil {
		log.Println("Error getting messages from this rooms:", flowjson.Err)
		return
	}
	c.toChannel(flowjson, Rows, flowjson.Message, flowjson.Users[0])
}

func (c *PostgresClient) GetMessagesFromNextRooms(flowjson *FlowJSON) {
	var arrayrooms []uint32
	for i := c.PaginationOffset; i < c.PaginationOffset+30; i++ {
		arrayrooms = append(arrayrooms, c.RoomsPagination[i])
	}
	var Rows pgx.Rows
	Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT r.room_id, m.message_id, m.payload, m.user_id
		FROM unnest($1) AS r(room_id)
		LEFT JOIN LATERAL (
			SELECT message_id, payload, user_id, timestamp
			FROM messages
			WHERE messages.room_id = r.room_id
			ORDER BY timestamp DESC
			LIMIT 30
		) AS m ON true
		WHERE r.room_id NOT IN ($2)
		ORDER BY r.room_id`, arrayrooms, flowjson.Rooms)
	if flowjson.Err != nil {
		log.Println("Error getting messages from this rooms:", flowjson.Err)
		return
	}
	c.toChannel(flowjson, Rows, flowjson.Message_id, flowjson.Message, flowjson.Users[0])
	if flowjson.Err == nil {
		c.PaginationOffset += 30
	}
}

func (c *PostgresClient) GetRoomUsersInfo(flowjson *FlowJSON) {
	var user_id int
	var name string
	var Rows pgx.Rows
	Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT u.user_id,u.name
		FROM users u JOIN room_users_info ru ON ru.user_id = u.user_id
		WHERE ru.room_id = $1`, c.UserID)
	if flowjson.Err != nil {
		log.Println("Error getting room info", flowjson.Err)
		return
	}
	c.toChannel(flowjson, Rows, &user_id, &name)
}

func (c *PostgresClient) FindUsers(flowjson *FlowJSON) {
	var user_id int
	var name string
	var Rows pgx.Rows
	Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT user_id,name FROM users WHERE name ILIKE $1 LIMIT 20`, flowjson.Name+"%")
	c.toChannel(flowjson, Rows, &user_id, &name)
}
func (c *PostgresClient) toChannel(flowjson *FlowJSON, rows pgx.Rows, dest ...any) {
	err := rows.Err() // checking for query timeout
	if err != nil {
		flowjson.Err = err
		return
	}
	for rows.Next() {
		err := rows.Scan(dest)
		if err != nil {
			log.Println("Error scanning rows:", err)
			flowjson.Err = err
			return
		}
		c.Chan <- *flowjson
	}
}
