package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var messageinsert = `INSERT INTO messages (payload,user_id,room_id) VALUES ($1,$2,$3)`

var workers = make(Workers)

// worker for safely creating private rooms for two people
type worker struct {
	Mutex  *sync.Mutex
	roomid uint64
}

type Workers map[string]worker

func (ws Workers) GetWorker(user1, user2 uint64) worker {
	ids := []uint64{user1, user2}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	key := strconv.FormatUint(user1, 10) + strconv.FormatUint(user2, 10)

	if w, ok := ws[key]; ok {
		return w
	}
	w := &sync.Mutex{}
	ws[key] = worker{
		Mutex: w,
	}
	return ws[key]
}

type FlowJSON struct {
	Mode      string   `json:"Mode"`
	Message   string   `json:"Message"`
	Users     []uint64 `json:"Users"`
	UsersInfo []uint64 `json:"UsersInfo"`
	Room      uint64   `json:"Room"`
	Name      string   `json:"Name"`
	Offset    string   `json:"Offset"`
	Status    string   `json:"Status"`
	Rows      pgx.Rows
	Tx        pgx.Tx
	Err       error
}

type PostgresClient struct {
	Sub            string
	Name           string
	UserID         uint64
	Conn           *pgxpool.Conn
	Rooms          map[uint32][]uint64 //  room id of group chat with user ids
	DuoRoomUsers   map[uint64]bool     // user id of private chat
	SearchUserList map[uint64]bool     // search user list with user id
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
	var userid uint64
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
	}
	if flowjson.Err != nil {
		flowjson.Status = "bad"
		flowjson.Err = flowjson.Tx.Rollback(context.Background())
		if flowjson.Err != nil {
			log.Println("ATTENTION Error rolling back transaction:", flowjson.Err)
		}
	}
}

// transaction insert messages plus increment unread messages in room_users table
func (c *PostgresClient) SendMessage(flowjson *FlowJSON) {
	_, flowjson.Err = flowjson.Tx.Exec(context.Background(), messageinsert, flowjson.Message, flowjson.Users[0], flowjson.Room)
	if flowjson.Err != nil {
		log.Println("Error inserting message:", flowjson.Err)
		return
	}
}

func (c *PostgresClient) CreateDuoRoom(flowjson *FlowJSON) {
	worker := workers.GetWorker(flowjson.Users[0], flowjson.Users[1])
	worker.Mutex.Lock()
	defer worker.Mutex.Unlock()
	if _, ok := c.DuoRoomUsers[flowjson.Users[1]]; ok {
		flowjson.Room = worker.roomid
		c.SendMessage(flowjson)
	}
	c.CreateRoom(flowjson)
}
func (c *PostgresClient) CreateRoom(flowjson *FlowJSON) {
	isGroup := false
	var roomID uint64
	var stmt *pgconn.StatementDescription
	tx := flowjson.Tx

	flowjson.Err = tx.QueryRow(context.Background(), "INSERT INTO rooms (name) VALUES ($1) RETURNING room_id", flowjson.Name).Scan(&roomID)
	if flowjson.Err != nil {
		log.Println("Error inserting room:", flowjson.Err)
		return
	}
	flowjson.Room = roomID
	stmt, flowjson.Err = tx.Prepare(context.Background(), "insert", "INSERT INTO room_users_info (user_id, room_id, is_group) VALUES ($1, $2, $3)")
	if flowjson.Err != nil {
		fmt.Println("Error preparing statement:", flowjson.Err)
		return
	}
	if flowjson.Name != "" {
		isGroup = true
	}
	for _, i := range flowjson.Users {
		_, flowjson.Err = tx.Exec(context.Background(), stmt.SQL, flowjson.Users[i], roomID, isGroup)
		if flowjson.Err != nil {
			log.Println("Error inserting room_users_info:", flowjson.Err)
			return
		}
	}
}

func (c *PostgresClient) GetMessages(flowjson *FlowJSON) {
	if flowjson.Room != 0 {
		flowjson.Rows, flowjson.Err = c.Conn.Query(context.Background(),
			`SELECT m.payload, m.user_id, m.room_id
			FROM messages m
			WHERE m.room_id = $1
			ORDER BY m.timestamp
			LIMIT 100 OFFSET &2`, flowjson.Room, flowjson.Offset)
	} else {
		flowjson.Rows, flowjson.Err = c.Conn.Query(context.Background(),
			`SELECT m.payload, m.user_id, m.room_id,r.room_user_info_id,r.is_group
			FROM messages m
			JOIN (
				SELECT room_id,room_user_info_id,is_group
				FROM room_users_info
				WHERE user_id = 1
			) r ON m.room_id = r.room_id
			ORDER BY m.room_id,m.timestamp LIMIT 100`, c.UserID)
	}
}

func (c *PostgresClient) GetRoomUsersInfo(flowjson *FlowJSON) {
	flowjson.Rows, flowjson.Err = c.Conn.Query(context.Background(),
		`SELECT room_id,room_user_info_id,unread
		FROM room_users_info
		WHERE user_id = 1 ORDER BY room_id`, c.UserID)
}

/*
///////////////TEST QUERIES/////////////////////


CREATE OR REPLACE FUNCTION notify_new_message()
RETURNS TRIGGER AS $$
DECLARE
  channel TEXT;
  user_id INTEGER;
BEGIN

  user_id := (SELECT user_id FROM room_users WHERE user_id = NEW.room_users_id);

  channel := 'user_' || user_id;

  PERFORM pg_notify(channel, 'new message in room: ' || NEW.room_id);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER new_message_trigger
AFTER INSERT ON messages
FOR EACH ROW
EXECUTE FUNCTION notify_new_message();




	SELECT * FROM rooms WHERE

INSERT INTO room_users (room_id, user_id, unread)
VALUES (3, 2, 5)
INSERT INTO users (subject,name) VALUES ('111nhudgokdhgfdidxfd','OKAYEG')

//////////////////

INSERT INTO rooms DEFAULT VALUES
/
INSERT INTO users (subject,name) VALUES ('dhdf982hfireb','aboba')
/
INSERT INTO messages (payload,user_id,room_id) VALUES ('hi arolfix',6,14)
/
INSERT INTO room_users_info (room_id,user_id,unread,is_group) VALUES ($1,$2,$3,$4)
/
SELECT payload,user_id,room_id FROM messages WHERE room_id IN (SELECT room_id FROM room_users_info WHERE user_id=6) ORDER BY room_id,timestamp
/
SELECT user_id,name FROM users WHERE name LIKE '%%';
/
CREATE SEQUENCE room_ids START 1;
/
INSERT INTO rooms (room_id) VALUES (nextval('room_ids')) RETURNING room_id
/
rooms_room_id_seq
/
BEGIN;
WITH roomval AS (INSERT INTO rooms DEFAULT VALUES RETURNING room_id)
INSERT INTO room_users_info (user_id, room_id, unread)
SELECT 1, room_id, 0 FROM roomval
UNION ALL
SELECT 5, room_id, 0 FROM roomval;
COMMIT;
//
SELECT payload,user_id FROM messages WHERE ((user_id = 6 OR to_user_id = 6) AND room_id IS NULL) ORDER BY timestamp
//
SELECT payload, user_id, room_id
FROM (
    SELECT payload, user_id, NULL AS room_id, timestamp
    FROM messages
    WHERE (user_id = 6 OR to_user_id = 6) AND room_id IS NULL
    UNION ALL
    SELECT m.payload, m.user_id, m.room_id, m.timestamp
    FROM messages m
    JOIN (
        SELECT room_id
        FROM room_users_info
        WHERE user_id = 6
    ) ru ON m.room_id = ru.room_id
)
ORDER BY COALESCE(room_id, user_id), timestamp;
//
SELECT m.payload, m.user_id, m.room_id
FROM messages m
JOIN (
	SELECT room_id
	FROM room_users_info
	WHERE user_id = 6
) ru ON m.room_id = ru.room_id
ORDER BY m.room_id,m.timestamp


//


BEGIN;


SELECT * FROM users WHERE user_id IN (1, 5) FOR UPDATE;


IF EXISTS (
  SELECT 1 FROM room_users_info ru1
  JOIN room_users_info ru2 ON ru1.room_id = ru2.room_id
  WHERE ru1.user_id = 1 AND ru2.user_id = 5 AND ru1.is_group = false
) THEN
  ROLLBACK;
  RAISE EXCEPTION 'room already exists';
END IF;


WITH roomval AS (INSERT INTO rooms (name) VALUES ("placeholder") RETURNING room_id)
INSERT INTO room_users_info (user_id, room_id, unread, is_group)
SELECT 1, room_id, 0, false FROM roomval
UNION ALL
SELECT 5, room_id, 0, false FROM roomval;
COMMIT;

*/
