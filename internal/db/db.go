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

var messageinsert = `INSERT INTO messages (payload,user_id,room_id) VALUES ($1,$2,$3)`

type FlowJSON struct {
	Mode       string   `json:"Mode"`
	Message    string   `json:"Message"`
	Users      []uint32 `json:"Users"`
	Rooms      []uint32 `json:"Room"`
	Name       string   `json:"Name"`
	Offset     string   `json:"Offset"`
	Status     string   `json:"Status"`
	SenderOnly bool
	Rows       pgx.Rows
	Tx         pgx.Tx
	Err        error
}

type FindUsersList struct {
	m map[uint32]bool //  room id of group chat with user ids
	sync.Mutex
}

type Rooms struct {
	m map[uint32]bool //  room id of group chat with user ids
	sync.Mutex
}
type PostgresClient struct {
	Sub           string
	Name          string
	UserID        uint32
	Conn          *pgxpool.Conn
	FindUsersList // current users
	Rooms
	RoomsPagination  []uint32
	RoomsCount       uint8 // no more than 250
	PaginationOffset uint8
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
func (c *PostgresClient) TxManage(flowjson *FlowJSON, fn func(*FlowJSON)) {
	if flowjson.SenderOnly {
		fn(flowjson)
		return
	}
	c.txBegin(flowjson)
	fn(flowjson)
	c.txCommit(flowjson)
}
func (c *PostgresClient) txBegin(flowjson *FlowJSON) {
	flowjson.Tx, flowjson.Err = flowjson.Tx.Begin(context.Background())
}
func (c *PostgresClient) txCommit(flowjson *FlowJSON) {
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
		flowjson.Err = flowjson.Tx.Rollback(context.Background())
		if flowjson.Err != nil {
			log.Println("ATTENTION Error rolling back transaction:", flowjson.Err)
		}
	}
}

// transaction insert messages plus increment unread messages in room_users table
func (c *PostgresClient) SendMessage(flowjson *FlowJSON) {
	//validating if room exists in our map
	if _, ok := c.Rooms.m[flowjson.Rooms[0]]; !ok {
		flowjson.Err = errors.New("room not found")
		return
	}

	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), messageinsert, flowjson.Message, c.UserID, flowjson.Rooms[0])
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
	var placeholder int
	flowjson.Err = c.Conn.QueryRow(context.Background(), `SELECT 1 FROM blocked_users 
	WHERE blocked_by_user_id = $1 AND blocked_user_id = $2`, c.UserID, flowjson.Users[1]).Scan(&placeholder)
	if flowjson.Err == nil {
		flowjson.Err = errors.New("user already blocked")
		return
	}
	var row = c.Conn.QueryRow(context.Background(), `SELECT room_id
		FROM duo_rooms_info
		WHERE user_id1 = $1 AND user_id2 = $2;`, c.UserID, flowjson.Users[1])
	flowjson.Err = row.Scan(&flowjson.Rooms[0])
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
	var placeholder int
	flowjson.Err = flowjson.Tx.QueryRow(context.Background(), `
		SELECT 1 FROM blocked_users 
		WHERE blocked_by_user_id = $1 AND blocked_user_id = $2`, c.UserID, flowjson.Users[1]).Scan(&placeholder)
	if flowjson.Err != nil {
		flowjson.Err = errors.New("this user is not blocked")
		return
	}
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), `DELETE FROM blocked_users 
			WHERE blocked_by_user_id = $1 AND blocked_user_id = $2`, c.UserID, flowjson.Users[1])
	if flowjson.Err != nil {
		log.Println("Error unblocking user", flowjson.Err)
		return
	}
}

// method for safely creating unique duo room
func (c *PostgresClient) CreateDuoRoom(flowjson *FlowJSON) {
	flowjson.Name = "private"
	if _, ok := c.FindUsersList.m[flowjson.Users[1]]; ok {
		var row = flowjson.Tx.QueryRow(context.Background(), `SELECT room_id
			FROM duo_rooms_info
			WHERE user_id1 = $1 AND user_id2 = $2;`, flowjson.Users[0], flowjson.Users[1])
		flowjson.Err = row.Scan(&flowjson.Rooms[0])
		if flowjson.Err != nil {
			c.CreateRoom(flowjson)
			_, flowjson.Err = flowjson.Tx.Exec(context.Background(), `INSERT INTO duo_rooms (user_id1, user_id2,room_id)`, flowjson.Users[0], flowjson.Users[1], flowjson.Rooms[0])
		} else {
			c.AddUsersToRoom(flowjson)
			return
		}

	} else {
		flowjson.Err = fmt.Errorf("user not found")
	}
}
func (c *PostgresClient) CreateRoom(flowjson *FlowJSON) {
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
	flowjson.Rooms[0] = roomID
	c.AddUsersToRoom(flowjson)
}
func (c *PostgresClient) AddUsersToRoom(flowjson *FlowJSON) {
	if flowjson.Mode == "AddUsersToRoom" {
		if _, ok := c.Rooms.m[flowjson.Rooms[0]]; !ok {
			flowjson.Err = errors.New("room not found")
			return
		}
	}
	var usercount uint8
	flowjson.Err = flowjson.Tx.QueryRow(context.Background(), `SELECT user_count FROM rooms WHERE room_id = $1 FOR UPDATE;`, flowjson.Rooms[0]).Scan(&usercount)
	if flowjson.Err != nil {
		log.Println("Error retrieving user count:", flowjson.Err)
		return
	}
	if len(flowjson.Users)+int(usercount) > 10 {
		flowjson.Err = errors.New("too many users in room")
		return
	}
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), `UPDATE rooms SET user_count = user_count + $1 WHERE room_id = ANY($2)`, len(flowjson.Users), flowjson.Rooms[0])
	if flowjson.Err != nil {
		log.Println("Error updating user count in room:", flowjson.Err)
		return
	}
	query := `
			INSERT INTO room_users_info (user_id, room_id)
			SELECT users_to_add.user_id, $1
			FROM (SELECT unnest($2) AS user_id) AS users_to_add
			JOIN users u ON u.user_id = users_to_add.user_id AND %s
			JOIN rooms r ON r.room_id = $1 AND r.isgroup = $3
			LEFT JOIN blocked_users bu ON bu.blocked_by_user_id = $4 AND bu.blocked_user_id = users_to_add.user_id
			WHERE bu.blocked_by_user_id IS NULL;
	`
	var condition string
	var isgroup bool
	if flowjson.Mode != "createDuoRoom" {
		condition = "u.allow_group_invites = true"
		isgroup = true
	} else {
		condition = "u.allow_direct_messages = true"
	}

	query = fmt.Sprintf(query, condition)

	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), query, flowjson.Rooms[0], flowjson.Users, isgroup, c.UserID)

}
func (c *PostgresClient) DeleteUsersFromRoom(flowjson *FlowJSON) {
	if flowjson.Mode == "DeleteUsersFromRoom" {
		if _, ok := c.Rooms.m[flowjson.Rooms[0]]; !ok {
			flowjson.Err = errors.New("room not found")
			return
		}
	}
	query := `DELETE FROM room_users_info
	WHERE room_id IN (
		SELECT room_id
		FROM rooms 
		WHERE room_id = $1 %s AND isgroup = $2
	) 
	AND user_id = ANY($3);`
	var condition string
	var isgroup bool
	ownerstr := fmt.Sprintf("AND owner = %s", fmt.Sprint(flowjson.Rooms[0]))
	if flowjson.Mode != "createDuoRoom" {
		condition = ownerstr
		isgroup = true
	} else {
		condition = ""
	}
	query = fmt.Sprintf(query, condition)
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), query, flowjson.Rooms[0], isgroup, flowjson.Users)
}

func (c *PostgresClient) GetAllRooms(flowjson *FlowJSON) {
	flowjson.Rows, flowjson.Err = flowjson.Tx.Query(context.Background(),
		`SELECT r.room_id
		FROM room_users_info ru JOIN rooms r ON ru.room_id = r.room_id
		WHERE ru.user_id = $1 AND is_visible = true 
		ORDER BY r.last_activity DESC;
		`, c.UserID)

	if flowjson.Err != nil {
		log.Println("Error getting all rooms:", flowjson.Err)
		return
	}
	defer c.Rooms.Mutex.Unlock()
	c.Rooms.Mutex.Lock()
	var roomID uint32
	for flowjson.Rows.Next() {
		flowjson.Rows.Scan(&roomID)
		c.Rooms.m[roomID] = false
		c.RoomsPagination = append(c.RoomsPagination, roomID)
	}
}

// load messages from a room
func (c *PostgresClient) GetMessagesFromRoom(flowjson *FlowJSON) {
	var payload string
	var user_id int
	flowjson.Rows, flowjson.Err = flowjson.Tx.Query(context.Background(),
		`SELECT payload,user_id,
		FROM messages 
		WHERE room_id = $1 AND message_id < $2
		ORDER BY message_id DESC`, flowjson.Rooms[0], flowjson.Offset)
	for flowjson.Rows.Next() {
		flowjson.Rows.Scan(&user_id, &payload)
		// send in channel here
	}
}

func (c *PostgresClient) GetMessagesFromNextRooms(flowjson *FlowJSON) {
	var room_id int
	var message_id int
	var payload string
	var user_id int

	c.PaginationOffset += 30
	flowjson.Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT r.room_id,m.message_id, m.payload, m.user_id
		FROM unnest($1) AS r(room_id)
		LEFT JOIN LATERAL (
			SELECT message_id, payload, user_id, timestamp
			FROM messages
			WHERE messages.room_id = r.room_id
			ORDER BY timestamp DESC
			LIMIT 30
		) AS m ON true
		ORDER BY r.room_id`, flowjson.Rooms)
	for flowjson.Rows.Next() {
		flowjson.Rows.Scan(&user_id, &message_id, &payload, &room_id)
		// send in channel here
	}
}

func (c *PostgresClient) GetRoomUsersInfo(flowjson *FlowJSON) {
	var user_id int
	var name string
	flowjson.Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT u.user_id,u.name
		FROM users u JOIN room_users_info ru ON ru.user_id = u.user_id
		WHERE ru.room_id = $1`, c.UserID)
	for flowjson.Rows.Next() {
		flowjson.Rows.Scan(&user_id, &name)
		// send in channel here
	}
}

func (c *PostgresClient) FindUsers(flowjson *FlowJSON) {
	var user_id int
	var name string
	flowjson.Rows, flowjson.Err = flowjson.Tx.Query(context.Background(),
		`SELECT user_id,name FROM users WHERE name ILIKE $1 LIMIT 20`, flowjson.Name+"%")
	for flowjson.Rows.Next() {
		flowjson.Rows.Scan(&user_id, &name)
		// send in channel here
	}
}
