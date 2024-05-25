package db

import (
	"context"
	"errors"
	"log"
	"strings"
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

func NewUser(sub, username string, pool *pgxpool.Pool) error {
	if strings.ContainsAny(username, " \t\n") {
		return errors.New("contains spaces")
	}
	_, err := pool.Exec(context.Background(), "INSERT INTO users (subject,username,allow_group_invites,allow_direct_messages) VALUES ($1,$2,true,true)", sub, username)
	return err
}

func ConnectToDB(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, connString)
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

func (c *PostgresClient) Channel() <-chan FlowJSON {
	return c.Chan
}
