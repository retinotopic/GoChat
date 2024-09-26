package db

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/internal/models"
)

type Message struct {
	Message   string `json:"Message"`
	MessageId uint32 `json:"MessageId" `
	RoomId    uint32 `json:"RoomId" `
	UserId    uint32 `json:"UserId" `
}

func SendMessage(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	m := &Message{}
	err := json.Unmarshal(event.Data, m)
	if err != nil {
		return err
	}
	if len(m.Message) == 0 || m.RoomId == 0 {
		return fmt.Errorf("malformed json")
	}
	_, err = tx.Exec(ctx, `INSERT INTO messages (message,user_id,room_id) VALUES ($1,$2,$3)`, m.Message, event.UserId, m.RoomId)
	if err != nil {
		return err
	}
	return err
}
func GetMessagesFromRoom(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	m := &Message{}
	err := json.Unmarshal(event.Data, m)
	if err != nil {
		return err
	}
	if m.MessageId == 0 || m.RoomId == 0 {
		return fmt.Errorf("malformed json")
	}
	rows, err := tx.Query(ctx,
		`SELECT message,user_id,
		FROM messages 
		WHERE room_id = $1 AND message_id < $2
		ORDER BY message_id DESC`, m.RoomId, m.MessageId)
	if err != nil {
		return err
	}
	resp, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[Message])
	if err != nil {
		return err
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	return err
}
