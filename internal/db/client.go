package db

import (
	"context"
	"fmt"
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

	if flowjson.Err != nil {
		log.Println("Error getting all rooms:", flowjson.Err)
		flowjson.Err = fmt.Errorf("error getting all rooms")
		return
	}
	for Rows.Next() {
		Rows.Scan(&flowjson.Rooms[0])
		c.RoomsPagination = append(c.RoomsPagination, flowjson.Rooms[0])
		c.Chan <- *flowjson
	}
}

// load messages from a room
func (c *PostgresClient) GetMessagesFromRoom(flowjson *FlowJSON) {
	var payload string
	var user_id int
	var Rows pgx.Rows
	Rows, flowjson.Err = flowjson.Tx.Query(context.Background(),
		`SELECT payload,user_id,
		FROM messages 
		WHERE room_id = $1 AND message_id < $2
		ORDER BY message_id DESC`, flowjson.Rooms[0], flowjson.Message_id)
	if flowjson.Err != nil {
		log.Println("Error getting messages from this rooms:", flowjson.Err)
		flowjson.Err = fmt.Errorf("error getting messages from this rooms")
		return
	}
	for Rows.Next() {
		Rows.Scan(&payload, &user_id)
		c.Chan <- *flowjson
	}
}

func (c *PostgresClient) GetMessagesFromNextRooms(flowjson *FlowJSON) {
	var room_id int
	var message_id int
	var payload string
	var user_id int
	var Rows pgx.Rows
	//slice of uint32
	var arrayrooms []uint32
	for i := c.PaginationOffset; i < c.PaginationOffset+30; i++ {
		arrayrooms = append(arrayrooms, c.RoomsPagination[i])
	}

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
		flowjson.Err = fmt.Errorf("error getting messages from next rooms")
		return
	}
	c.PaginationOffset += 30
	for Rows.Next() {
		err := Rows.Scan(&room_id, &message_id, &payload, &user_id)
		if err != nil {
			log.Fatalln("Error scanning rows:", err)
			return
		}
		c.Chan <- *flowjson
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
	for Rows.Next() {
		Rows.Scan(&user_id, &name)
		c.Chan <- *flowjson
	}
}

func (c *PostgresClient) FindUsers(flowjson *FlowJSON) {
	var user_id int
	var name string
	var Rows pgx.Rows
	Rows, flowjson.Err = flowjson.Tx.Query(context.Background(),
		`SELECT user_id,name FROM users WHERE name ILIKE $1 LIMIT 20`, flowjson.Name+"%")
	for Rows.Next() {
		Rows.Scan(&user_id, &name)
		c.Chan <- *flowjson
	}
}
