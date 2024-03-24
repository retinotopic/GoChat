package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var messageinsert = `INSERT INTO messages (payload,user_id,room_id) VALUES ($1,$2,$3)`

var workers = make(Workers)

// worker for safely creating private rooms for two people

type Workers map[string]*sync.Mutex

func (ws Workers) GetWorker(user1, user2 uint32) *sync.Mutex {
	ids := []uint32{user1, user2}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	key := fmt.Sprintf("%d%d", user1, user2)

	if w, ok := ws[key]; ok {
		return w
	}
	w := &sync.Mutex{}
	ws[key] = w
	return ws[key]
}

type FlowJSON struct {
	Mode    string   `json:"Mode"`
	Message string   `json:"Message"`
	Users   []uint32 `json:"Users"`
	Room    uint32   `json:"Room"`
	Name    string   `json:"Name"`
	Offset  string   `json:"Offset"`
	Status  string   `json:"Status"`
	Rows    pgx.Rows
	Tx      pgx.Tx
	Mutex   *sync.Mutex
	Err     error
}

type UsersToRooms struct {
	m map[uint32]bool //  room id of group chat with user ids
	sync.Mutex
}
type PostgresClient struct {
	Sub          string
	Name         string
	UserID       uint32
	Conn         *pgxpool.Conn
	UsersToRooms // search user list with user id
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
	return &PostgresClient{
		Sub:    sub,
		Conn:   conn,
		UserID: userid,
		Name:   name,
	}, nil
}
func (c *PostgresClient) TxBegin(flowjson *FlowJSON) {
	flowjson.Tx, flowjson.Err = c.Conn.Begin(context.Background())
}
func (c *PostgresClient) TxCommit(flowjson *FlowJSON) {
	if flowjson.Err == nil {
		flowjson.Err = flowjson.Tx.Commit(context.Background())
		if flowjson.Err != nil {
			flowjson.Status = "bad"
			flowjson.Err = flowjson.Tx.Rollback(context.Background())
			if flowjson.Err != nil {
				log.Println("ATTENTION Error rolling back transaction:", flowjson.Err)
			}
		}
	} else {
		flowjson.Status = "bad"
	}
}

// transaction insert messages plus increment unread messages in room_users table
func (c *PostgresClient) SendMessage(flowjson *FlowJSON) {
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), messageinsert, flowjson.Message, c.UserID, flowjson.Room)
	if flowjson.Err != nil {
		log.Println("Error inserting message:", flowjson.Err)
		return
	}
	// increment unread messages in room_users table
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), "UPDATE room_users_info SET unread=unread+1,isVisible = true WHERE room_id = $1", flowjson.Room)
	if flowjson.Err != nil {
		fmt.Println("Error preparing statement:", flowjson.Err)
		return
	}
}

// method for set unread to 0
func (c *PostgresClient) MarkAsRead(flowjson *FlowJSON) {
	_, flowjson.Err = c.Conn.Query(context.Background(), "UPDATE room_users_info SET unread=0 WHERE room_user_info = $1", flowjson.Users[0])
	if flowjson.Err != nil {
		log.Println("Error inserting room_users_info:", flowjson.Err)
	}
}

// method for safely creating unique duo room
func (c *PostgresClient) CreateDuoRoom(flowjson *FlowJSON) {
	worker := workers.GetWorker(flowjson.Users[0], flowjson.Users[1])
	flowjson.Mutex = worker
	flowjson.Mutex.Lock()
	flowjson.Name = "private"
	if _, ok := c.UsersToRooms.m[flowjson.Users[1]]; ok {
		var rows pgx.Rows
		rows, flowjson.Err = c.Conn.Query(context.Background(), `SELECT user_id1,user_id2
			FROM duo_rooms_info
			WHERE user_id1 = $1 AND user_id2 = $2;`, flowjson.Users[0], flowjson.Users[1])

		if !rows.Next() {
			c.CreateRoom(flowjson)

		} else {
			flowjson.Err = fmt.Errorf("room already exists")
			return
		}

	} else {
		flowjson.Err = fmt.Errorf("user not found")
	}
}
func (c *PostgresClient) CreateRoom(flowjson *FlowJSON) {
	// if len of users is more than 10 issue a error
	if len(flowjson.Users) > 10 {
		flowjson.Err = errors.New("too many users")
		return
	}
	var roomID uint32
	var isDuoRoom bool
	if flowjson.Mode == "createDuoRoom" {
		isDuoRoom = true
	}
	// create new room and return room id
	flowjson.Err = flowjson.Tx.QueryRow(context.Background(), "INSERT INTO rooms (name,isDuoRoom) VALUES ($1,$2) RETURNING room_id", flowjson.Name, isDuoRoom).Scan(&roomID)
	if flowjson.Err != nil {
		log.Println("Error inserting room:", flowjson.Err)
		return
	}
	flowjson.Room = roomID
	c.AddUsersToRoom(flowjson)
}
func (c *PostgresClient) AddUsersToRoom(flowjson *FlowJSON) {
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), `INSERT INTO room_users_info (user_id, room_id)
	SELECT users_to_add.user_id, $1
	FROM (SELECT unnest($2) AS user_id) AS users_to_add
	JOIN users u ON u.user_id = users_to_add.user_id AND u.allow_group_invites = true
	JOIN rooms r ON r.room_id = $1 AND r.isgroup = true
	LEFT JOIN blocked_users bu ON bu.blocked_by_user_id = $3 AND bu.blocked_user_id = users_to_add.user_id
	WHERE bu.blocked_by_user_id IS NULL;`, flowjson.Room, flowjson.Users, c.UserID)
}
func (c *PostgresClient) DeleteUsersFromRoom(flowjson *FlowJSON) {
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), `
	DELETE FROM room_users_info
	WHERE room_id = $1
	  AND EXISTS (
		SELECT 1
		FROM (SELECT unnest($2) AS user_id) AS users_to_remove
		JOIN users u ON u.user_id = users_to_remove.user_id
		JOIN rooms r ON r.room_id = $1 AND r.isgroup = true AND r.owner = $3
		WHERE room_users_info.user_id = users_to_remove.user_id
	  );`, flowjson.Room, flowjson.Users, c.UserID)
}
func (c *PostgresClient) GetTopMessages(flowjson *FlowJSON) {
	flowjson.Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT r.room_id, m.payload, m.user_id, m.timestamp
		FROM (
			SELECT room_id
			FROM room_users_info
			WHERE user_id = $1 AND is_visible = true LIMIT 30 OFFSET $2
		) AS r
		LEFT JOIN LATERAL (
			SELECT payload, user_id, timestamp
			FROM messages
			WHERE messages.room_id = r.room_id
			ORDER BY timestamp DESC
			LIMIT 30
		) AS m ON true
		ORDER BY r.room_id`, c.UserID, flowjson.Offset)
}

// load messages for a room or last 100 messages for all rooms
func (c *PostgresClient) GetMessagesFromRoom(flowjson *FlowJSON) {
	flowjson.Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT m.payload, m.user_id, m.room_id
		FROM messages m
		WHERE m.room_id = $1
		ORDER BY m.timestamp
		LIMIT 100 OFFSET $2`, flowjson.Room, flowjson.Offset)
}

func (c *PostgresClient) GetRoomUsersInfo(flowjson *FlowJSON) {
	flowjson.Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT room_id,room_user_info_id,user_id
		FROM room_users_info
		WHERE user_id = $1 ORDER BY room_id`, c.UserID)
}
