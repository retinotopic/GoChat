package db

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/retinotopic/GoChat/pkg/str"
)

type FlowJSON struct {
	Mode       string   `json:"Mode"`
	Message    string   `json:"Message"`
	Users      []uint32 `json:"Users"`
	Room       uint32   `json:"Room"`
	Name       string   `json:"Name"`
	Message_id string   `json:"Offset"`
	Status     string   `json:"Status"`
	Bool       bool     `json:"Bool"`
	Tx         pgx.Tx
	Err        error
}

var pool *pgxpool.Pool

func init() {
	var err error
	pool, err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("%v", err)
	}
}

type PostgresClient struct {
	Sub              string
	Name             string
	UserID           uint32
	Mutex            sync.Mutex
	RoomsPagination  []uint32
	RoomsCount       uint8 // no more than 250
	PaginationOffset uint8
	funcmap          map[string]fnAPI
	Chan             chan FlowJSON
}
type fnAPI struct {
	Fn         func(context.Context, *FlowJSON)
	ClientOnly bool
}

func NewUser(ctx context.Context, sub, username string) error {
	if strings.ContainsAny(username, " \t\n") {
		return errors.New("contains spaces")
	}
	_, err := pool.Exec(ctx, "INSERT INTO users (subject,username,allow_group_invites,allow_direct_messages) VALUES ($1,$2,true,true)", sub, username)
	return err
}

func ConnectToDB(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	return db, ctx.Err()
}
func NewClient(ctx context.Context, sub string) (*PostgresClient, error) {
	// check if user exists
	row := pool.QueryRow(ctx, "SELECT user_id,username FROM users WHERE subject=$1", sub)
	var name string
	var userid uint32
	err := row.Scan(&userid, &name)
	if err != nil {
		return nil, err
	}
	pc := &PostgresClient{
		Sub:    sub,
		UserID: userid,
		Name:   name,
		Chan:   make(chan FlowJSON, 1000),
	}
	pc.funcmap = map[string]fnAPI{
		"GetAllRooms":         {pc.GetAllRooms, true},
		"GetMessagesFromRoom": {pc.GetMessagesFromRoom, true},
		"GetNextRooms":        {pc.GetNextRooms, true},
		"FindUsers":           {pc.FindUsers, true},
		"SendMessage":         {pc.SendMessage, false},
		"CreateDuoRoom":       {pc.CreateDuoRoom, false},
		"CreateGroupRoom":     {pc.CreateRoom, false},
		"AddUsersToRoom":      {pc.AddUsersToRoom, false},
		"DeleteUsersFromRoom": {pc.DeleteUsersFromRoom, false},
		"BlockUser":           {pc.BlockUser, false},
		"UnblockUser":         {pc.UnblockUser, false},
		"ChangeUsername":      {pc.ChangeUsername, false},
		"ChangePrivacyDirect": {pc.ChangePrivacyDirect, false},
		"ChangePrivacyGroup":  {pc.ChangePrivacyGroup, false},
	}
	return pc, nil
}

// transaction insert messages
func (c *PostgresClient) SendMessage(ctx context.Context, flowjson *FlowJSON) {
	_, flowjson.Err = pool.Exec(ctx, `INSERT INTO messages (payload,user_id,room_id) VALUES ($1,$2,$3)`, flowjson.Message, c.UserID, flowjson.Room)
	if flowjson.Err != nil {
		log.Println("Error inserting message:", flowjson.Err)
		return
	}
}
func (c *PostgresClient) ChangeUsername(ctx context.Context, flowjson *FlowJSON) {
	username := str.NormalizeString(flowjson.Name)
	_, flowjson.Err = pool.Exec(ctx, "UPDATE users SET username = $1 WHERE user_id = $2", username, c.UserID)
	if flowjson.Err != nil {
		log.Println("Error changing username", flowjson.Err)
		return
	}
}
func (c *PostgresClient) ChangePrivacyDirect(ctx context.Context, flowjson *FlowJSON) {
	_, flowjson.Err = pool.Exec(ctx, "UPDATE users SET allow_direct_messages = $1 WHERE user_id = $2", flowjson.Bool, c.UserID)
	if flowjson.Err != nil {
		log.Println("Error changing username", flowjson.Err)
		return
	}
}
func (c *PostgresClient) ChangePrivacyGroup(ctx context.Context, flowjson *FlowJSON) {
	_, flowjson.Err = pool.Exec(ctx, "UPDATE users SET allow_group_invites = $1 WHERE user_id = $2", flowjson.Bool, c.UserID)
	if flowjson.Err != nil {
		log.Println("Error changing username", flowjson.Err)
		return
	}
}

// Blocking user and delete user from duo room
func (c *PostgresClient) BlockUser(ctx context.Context, flowjson *FlowJSON) {

	c.IsDuoRoomExist(ctx, flowjson)
	if flowjson.Err != nil {
		log.Println(flowjson.Err, "isroomexist err")
		return
	}
	if flowjson.Room != 0 {
		c.DeleteUsersFromRoom(ctx, flowjson)
		if flowjson.Err != nil {
			log.Println(flowjson.Err, "DeleteUsersFromRoom err")
			return
		}
	}

	_, flowjson.Err = flowjson.Tx.Exec(ctx, `INSERT INTO blocked_users (blocked_by_user_id, blocked_user_id)
		VALUES ($1, $2)`, flowjson.Users[0], flowjson.Users[1])
	if flowjson.Err != nil {
		log.Println("Error blocking user", flowjson.Err)
		return
	}
}

// Unblocking user
func (c *PostgresClient) UnblockUser(ctx context.Context, flowjson *FlowJSON) {
	_, flowjson.Err = flowjson.Tx.Exec(ctx, `DELETE FROM blocked_users 
			WHERE blocked_by_user_id = $1 AND blocked_user_id = $2`, c.UserID, flowjson.Users[1])
	if flowjson.Err != nil {
		log.Println("Error unblocking user", flowjson.Err)
		return
	}
}

func (c *PostgresClient) Channel() <-chan FlowJSON {
	return c.Chan
}
func (c *PostgresClient) ClearChannel() {
	for len(c.Chan) > 0 {
		<-c.Chan
	}
}
