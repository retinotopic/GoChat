package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FlowJSON struct {
	Mode       string   `json:"Mode"`
	Message    string   `json:"Message"`
	Users      []uint32 `json:"Users"`
	Rooms      []uint32 `json:"Room"`
	Name       string   `json:"Name"`
	Message_id string   `json:"Offset"`
	Status     string   `json:"Status"`
	SenderOnly bool
	Tx         pgx.Tx
	Err        error
}

type PostgresClient struct {
	Sub              string
	Name             string
	UserID           uint32
	Conn             *pgxpool.Conn
	Mutex            sync.Mutex
	RoomsPagination  []uint32
	RoomsCount       uint8 // no more than 250
	PaginationOffset uint8
	funcmap          map[string]func(*FlowJSON)
	Chan             chan FlowJSON
}

func ConnectToDB(connString string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	return db, err
}
func NewClient(sub string, pool *pgxpool.Pool) (*PostgresClient, error) {
	// check if user exists
	row, err := pool.Query(context.Background(), "SELECT * FROM users WHERE subject=$1", sub)
	if err != nil {
		return nil, err
	}
	var name string
	var userid uint32
	err = row.Scan(&name, &userid)
	if err != nil {
		return nil, err
	}
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	pc := &PostgresClient{
		Sub:    sub,
		Conn:   conn,
		UserID: userid,
		Name:   name,
		Chan:   make(chan FlowJSON, 10),
	}
	pc.funcmap = map[string]func(*FlowJSON){
		"SendMessage":              pc.SendMessage,
		"GetAllRooms":              pc.GetAllRooms,
		"CreateDuoRoom":            pc.CreateDuoRoom,
		"CreateGroupRoom":          pc.CreateRoom,
		"GetMessagesFromNextRooms": pc.GetMessagesFromNextRooms,
		"AddUsersToRoom":           pc.AddUsersToRoom,
		"DeleteUsersFromRoom":      pc.DeleteUsersFromRoom,
		"BlockUser":                pc.BlockUser,
		"UnblockUser":              pc.UnblockUser,
	}
	return pc, nil
}

// transaction insert messages plus increment unread messages in room_users table
func (c *PostgresClient) SendMessage(flowjson *FlowJSON) {
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), `INSERT INTO messages (payload,user_id,room_id) VALUES ($1,$2,$3)`, flowjson.Message, c.UserID, flowjson.Rooms[0])
	if flowjson.Err != nil {
		log.Println("Error inserting message:", flowjson.Err)
		return
	}
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), `UPDATE rooms
		SET last_activity = NOW()
		WHERE room_id = $1 AND NOW() - last_activity > INTERVAL '24 hours'`, flowjson.Rooms[0])
	if flowjson.Err != nil {
		log.Println("Error inserting message:", flowjson.Err)
		return
	}
}

// blocking user and delete user from duo room
func (c *PostgresClient) BlockUser(flowjson *FlowJSON) {

	c.IsDuoRoomExist(flowjson)
	if flowjson.Err == nil {
		c.DeleteUsersFromRoom(flowjson)
	}

	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), `INSERT INTO blocked_users (blocked_by_user_id, blocked_user_id)
		VALUES ($1, $2)`, flowjson.Users[0], flowjson.Users[1])
	if flowjson.Err != nil {
		log.Println("Error blocking user", flowjson.Err)
		return
	}
}
func (c *PostgresClient) UnblockUser(flowjson *FlowJSON) {
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), `DELETE FROM blocked_users 
			WHERE blocked_by_user_id = $1 AND blocked_user_id = $2`, c.UserID, flowjson.Users[1])
	if flowjson.Err != nil {
		log.Println("Error unblocking user", flowjson.Err)
		return
	}
}

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
			JOIN rooms r ON r.room_id = $1 AND r.isgroup = $3
			LEFT JOIN blocked_users bu ON bu.blocked_by_user_id = $4 AND bu.blocked_user_id = users_to_add.user_id
			WHERE bu.blocked_by_user_id IS NULL;`
	var condition string
	var isgroup bool
	if flowjson.Mode == "createDuoRoom" {
		condition = "u.allow_direct_messages = true"
	} else {
		condition = "u.allow_group_invites = true"
		isgroup = true
	}

	query = fmt.Sprintf(query, condition)

	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), query, flowjson.Rooms[0], flowjson.Users, isgroup, c.UserID)

}
func (c *PostgresClient) DeleteUsersFromRoom(flowjson *FlowJSON) {
	if flowjson.Mode == "DeleteUsersFromRoom" {
		if len(flowjson.Users) != 1 || flowjson.Users[0] != c.UserID {
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
		WHERE room_id = $2 AND isgroup = $3
	);`
	var isgroup bool

	if flowjson.Mode != "BlockUser" {
		isgroup = true
	}
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), query, flowjson.Users, flowjson.Rooms[0], isgroup)
}

func (c *PostgresClient) Channel() <-chan FlowJSON {
	return c.Chan
}
