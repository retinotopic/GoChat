package db

import (
	"context"
	"errors"
	"math"
	"strconv"

	json "github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/server/models"
)

type Message struct {
	MessagePayload string `json:"MessagePayload"`
	MessageId      uint64 `json:"MessageId" `
	RoomId         uint64 `json:"RoomId" `
	UserId         uint64 `json:"UserId" `
}

func SendMessage(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	m := &Message{}
	err := json.Unmarshal(event.Data, m)
	if err != nil {
		return err
	}
	if len(m.MessagePayload) == 0 || m.RoomId == 0 {
		return errors.New("malformed json")
	}
	_, err = tx.Exec(ctx, `INSERT INTO messages (MessagePayload,user_id,room_id) VALUES ($1,$2,$3)`, m.MessagePayload, event.UserId, m.RoomId)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 1
	event.SubForPub = []string{strconv.Itoa(int(m.RoomId))}
	return err
}
func GetMessagesFromRoom(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	m := &Message{}
	err := json.Unmarshal(event.Data, m)
	if err != nil {
		return err
	}

	if m.RoomId == 0 {
		return errors.New("malformed json")
	}
	if m.MessageId == 0 {
		m.MessageId = math.MaxUint64
	}
	rows, err := tx.Query(ctx,
		`SELECT MessagePayload,user_id,message_id,room_id
		FROM messages 
		WHERE room_id = $1 AND message_id < $2
		ORDER BY message_id DESC LIMIT 40`, m.RoomId, m.MessageId)
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
