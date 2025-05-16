package db

import (
	"context"
	"errors"
	"strconv"

	json "github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/server/models"
)

func SendMessage(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	m, err := models.UnmarshalEvent[models.Message](event.Data)
	if err != nil {
		return err
	}
	if len(m.MessagePayload) == 0 || m.RoomId == 0 {
		return errors.New("malformed json")
	}
	if err := tx.QueryRow(ctx, `SELECT 1 FROM room_users_info WHERE room_id = $1 AND user_id = $2`, m.RoomId, event.UserId).Scan(new(int)); err != nil {
		err = errors.New("you have no permission to send messages to this room")
		return err
	}
	err = tx.QueryRow(ctx, `WITH msg AS (INSERT INTO messages (message_payload,user_id,room_id) VALUES ($1,$2,$3)
		returning user_id)
		SELECT u.user_name FROM msg m JOIN users u ON u.user_id = m.user_id`, m.MessagePayload, event.UserId, m.RoomId).Scan(&m.Username)
	if err != nil {
		return err
	}
	m.UserId = event.UserId
	event.Data, err = json.Marshal(m)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 1
	event.OrderCmd[1] = -1
	event.PublishChs = []string{"room" + strconv.Itoa(int(m.RoomId))}
	return err
}

func GetMessagesFromRoom(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	m, err := models.UnmarshalEvent[models.Message](event.Data)
	if err != nil {
		return err
	}
	getmsgs := getOldMessages
	if m.RoomId == 0 {
		return errors.New("malformed json")
	}
	if m.MessageId == 0 {
		getmsgs = getNewMessages
	}
	rows, err := tx.Query(ctx, getmsgs, m.RoomId, m.MessageId)
	if err != nil {
		return err
	}
	resp, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.Message])
	if err != nil {
		return err
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	return err
}
