package db

import (
	"context"
	"fmt"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/internal/models"
)

type Message struct {
	MessagePayload string `json:"MessagePayload"`
	MessageId      uint32 `json:"MessageId" `
	RoomId         uint32 `json:"RoomId" `
	UserId         uint32 `json:"UserId" `
}

func SendMessage(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	m := &Message{}
	err := json.Unmarshal(event.Data, m)
	if err != nil {
		return err
	}
	if len(m.MessagePayload) == 0 || m.RoomId == 0 {
		return fmt.Errorf("malformed json")
	}
	_, err = tx.Exec(ctx, `INSERT INTO messages (MessagePayload,user_id,room_id) VALUES ($1,$2,$3)`, m.MessagePayload, event.UserId, m.RoomId)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 1
	event.SubForPub = []string{strconv.Itoa(int(m.RoomId))}
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
		`SELECT MessagePayload,user_id,
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
