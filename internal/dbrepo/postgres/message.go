package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/internal/models"
)

type Message struct {
	Message   string `json:"Message"`
	MessageId uint32 `json:"MessageId" `
	RoomId    uint32 `json:"RoomId" `
	UserId    uint32 `json:"UserId" `
}
type Messages struct {
	Messages []Message `json:"Messages" `
}

func SendMessage(ctx context.Context, tx pgx.Tx, event *models.Event) error {

	_, err := tx.Exec(ctx, `INSERT INTO messages (message,user_id,room_id) VALUES ($1,$2,$3)`, event.Payload, event.UserId, event.PayloadArr[0])
	if err != nil {
		return err
	}
	return err
}
func GetMessagesFromRoom(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	rows, err := tx.Query(ctx,
		`SELECT message,user_id,
		FROM messages 
		WHERE room_id = $1 AND message_id < $2
		ORDER BY message_id DESC`, event.PayloadIds[0], event.PayloadIds[1])
	if err != nil {
		return err
	}
	r.Rooms, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[Message])
	if err != nil {
		return err
	}
	return err
}
