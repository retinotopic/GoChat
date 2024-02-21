package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClient struct {
	sub  string
	conn *pgxpool.Conn
}

func ConnectToDB(connString string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	return db, err

}
func NewClient(sub string, pool *pgxpool.Pool) (*PostgresClient, error) {
	// check if user exists
	_, err := pool.Exec(context.Background(), "SELECT * FROM users WHERE subject=$1", sub)
	if err != nil {
		return nil, err
	}
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	return &PostgresClient{
		sub:  sub,
		conn: conn,
	}, nil
}

// transaction insert messages plus increment unread messages in room_users table
func (c *PostgresClient) InsertMessage(room_id int, message string, room_user_id int) error {

	tx, err := c.conn.Begin(context.Background())
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
	_, err = tx.Conn().Exec(context.Background(), "INSERT INTO messages (room_id, message, room_user_id) VALUES ($1, $2, $3)", room_id, message, room_user_id)
	if err != nil {
		return err
	}
	_, err = c.conn.Exec(context.Background(), "UPDATE room_users SET unread=unread+1 WHERE room_id=$2 AND user_id=$3", 1, room_id, room_user_id)
	if err != nil {
		return err
	}
	return nil
}

/*



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

*/
