package db

import (
	"context"
	"os"
	"sort"
	"strconv"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// worker for creating thread safe private rooms for two people
type Worker struct {
	mutex sync.Mutex
	done  bool
}

func (w *Worker) CreateRoom(message string, user1, user2 int) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.done {
		return
	}
	//
	// CREATING ROOM HERE
	//
	w.done = true

}

var workers = make(Workers)

type Workers map[string]*Worker

func (ws Workers) GetWorker(user1, user2 int) *Worker {
	ids := []int{user1, user2}
	sort.Ints(ids)
	key := strconv.Itoa(ids[0]) + strconv.Itoa(ids[1])
	if w, ok := ws[key]; ok {
		return w
	}
	w := &Worker{}
	ws[key] = w
	return w
}

type PostgresClient struct {
	Sub            string
	Name           string
	UserID         uint64
	Conn           *pgxpool.Conn
	Rooms          map[uint32][]int //  room id of group chat with user ids
	DuoRoomUsers   map[uint32]bool  // user id of private chat
	SearchUserList map[uint32]bool  // search user list with user id
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

// transaction insert messages plus increment unread messages in room_users table
func (c *PostgresClient) SendMessage(payload string) error {
	room := 0
	_, err := c.Conn.Exec(context.Background(), "INSERT INTO messages (payload,user_id,room_id) VALUES ($1,$2,$3)", payload, c.UserID, room)
	return err
}
func (c *PostgresClient) CreateDuoRoom(data string) error {
	worker := workers.GetWorker(13, 4)
	worker.CreateRoom("message1", 13, 4)
	return nil
}
func (c *PostgresClient) CreateGroupRoom(data string) error {
	tx, err := c.Conn.Begin(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}()
	return nil
}
func (c *PostgresClient) GetMessages(room uint64, offset uint64) (pgx.Rows, error) {
	if room == 0 {
		rows, err := c.Conn.Query(context.Background(),
			`SELECT m.payload, m.user_id, m.room_id
			FROM messages m
			WHERE m.room_id = $1
			ORDER BY m.timestamp
			LIMIT 100 OFFSET &2`, room, offset)
		return rows, err
	} else {
		rows, err := c.Conn.Query(context.Background(),
			`SELECT m.payload, m.user_id, m.room_id
			FROM messages m
			JOIN (
				SELECT room_id
				FROM room_users_info
				WHERE user_id = $1
			) ru ON m.room_id = ru.room_id
			ORDER BY m.room_id,m.timestamp LIMIT 100`, c.UserID)
		return rows, err
	}
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
