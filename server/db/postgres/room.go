package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	json "github.com/bytedance/sonic"

	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/server/models"
)

func GetUserRooms(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) (err error) {
	rows, err := tx.Query(ctx,
		`SELECT ru.room_id,ru.user_id,r.room_name,r.is_group,r.created_by_user_id,u.user_name
		FROM room_users_info ru JOIN rooms r ON ru.room_id = r.room_id JOIN users u ON ru.user_id = u.user_id
		WHERE ru.room_id IN (SELECT room_id FROM room_users_info WHERE user_id = $1)
		ORDER BY r.last_activity DESC, r.room_id`, event.UserId)
	if err != nil {
		return err
	}
	resp, err := NormalizeRoom(rows, 0)
	if err != nil {
		return err
	}
	if len(resp) == 0 {
		return errors.New("no rooms found")
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	event.OrderCmd[1] = 3
	return err
}

// method for creating unique duo room
func CreateDuoRoom(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r, err := models.UnmarshalEvent[models.RoomRequest](event.Data)
	if err != nil {
		return err
	}
	r.RoomIds = make([]uint64, 1)
	if len(r.UserIds) == 0 {
		return errors.New("malformed json")
	}
	first, second := r.UserIds[0], event.UserId
	if event.UserId > r.UserIds[0] {
		first, second = event.UserId, r.UserIds[0]
	}
	r.UserIds = append(r.UserIds, event.UserId)
	userstr := models.ConvertUint64ToString(r.UserIds)

	row := tx.QueryRow(ctx, `SELECT room_id
		FROM duo_rooms
		WHERE (user_id1 = $1 AND user_id2 = $2) OR (user_id2 = $1 AND user_id1 = $2) FOR UPDATE`, first, second)

	err = row.Scan(&r.RoomIds[0])
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	r.RoomName = "Duo Room with users: " + userstr[0] + " and " + userstr[1]
	var rows pgx.Rows
	is_group := false
	if r.RoomIds[0] == 0 {
		err := tx.QueryRow(ctx, CreateRoom, r.RoomName, is_group, event.UserId).Scan(&r.RoomIds[0])
		if err != nil {
			return err
		}
		t, err := tx.Exec(ctx, `INSERT INTO duo_rooms (user_id1,user_id2,room_id) VALUES ($1,$2,$3)`, r.UserIds[1], r.UserIds[0], r.RoomIds[0])
		if t.RowsAffected() == 0 {
			return errors.New("no rows affected")
		}
		if err != nil {
			return err
		}
	}
	tag, err := tx.Exec(ctx, addUsersToRoomDirect, r.RoomIds[0], r.UserIds, false, event.UserId) // bool param is is_group
	if err != nil {
		return err
	}
	if int(tag.RowsAffected()) != len(r.UserIds) {
		return errors.New("no users added")
	}
	rows, err = tx.Query(ctx, getRoomUsers, r.RoomIds[0]) // bool param is is_group
	if err != nil {
		return err
	}
	resp, err := NormalizeRoom(rows, 0)
	if err != nil {
		return err
	}
	if len(resp) == 0 {
		return errors.New("no users added")
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 2
	event.OrderCmd[1] = 1
	event.UserChs = userstr
	event.PublishChs = []string{strconv.Itoa(int(r.RoomIds[0]))}
	event.Kind = true
	return err
}

func CreateGroupRoom(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r, err := models.UnmarshalEvent[models.RoomRequest](event.Data)
	if err != nil {
		return err
	}
	r.RoomIds = make([]uint64, 1)
	if len(r.RoomName) == 0 || len(r.UserIds) == 0 {
		return errors.New("malformed json")
	}
	r.UserIds = append(r.UserIds, event.UserId)
	userstr := models.ConvertUint64ToString(r.UserIds)
	// create new room and return room id
	err = tx.QueryRow(ctx, CreateRoom, r.RoomName, true, event.UserId).Scan(&r.RoomIds[0]) // bool param is is_group
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, addUsersToRoomGroup, r.RoomIds[0], r.UserIds, true, event.UserId) // bool param is is_group
	if err != nil {
		return err
	}
	if int(tag.RowsAffected()) != len(r.UserIds) {
		return errors.New("no users added")
	}
	rows, err := tx.Query(ctx, getRoomUsers, r.RoomIds[0]) // bool param is is_group
	if err != nil {
		return err
	}
	resp, err := NormalizeRoom(rows, 0)
	if err != nil {
		return err
	}
	if len(resp) == 0 {
		return errors.New("no users added")
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 2
	event.OrderCmd[1] = 1
	event.UserChs = userstr
	event.PublishChs = []string{strconv.Itoa(int(r.RoomIds[0]))}
	event.Kind = true
	return err
}
func AddUsersToRoom(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r, err := models.UnmarshalEvent[models.RoomRequest](event.Data)
	if err != nil {
		return err
	}
	if len(r.RoomIds) == 0 || len(r.UserIds) == 0 {
		return errors.New("malformed json")
	}
	if err := tx.QueryRow(ctx, `SELECT 1 FROM rooms WHERE room_id = $1 AND created_by_user_id = $2 FOR UPDATE`, r.RoomIds[0], event.UserId).Scan(new(int)); err != nil {
		err = errors.New("you have no permission to add users to this room")
		return err
	}
	tag, err := tx.Exec(ctx, addUsersToRoomGroup, r.RoomIds[0], r.UserIds, true, event.UserId) // bool param is is_group
	if err != nil {
		return err
	}
	if int(tag.RowsAffected()) != len(r.UserIds) {
		return errors.New("no users added")
	}
	rows, err := tx.Query(ctx, getRoomUsers, r.RoomIds[0]) // bool param is is_group
	if err != nil {
		return err
	}
	resp, err := NormalizeRoom(rows, 0)
	if err != nil {
		return err
	}
	if len(resp) == 0 {
		return errors.New("no users added")
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 2
	event.OrderCmd[1] = 1
	event.UserChs = models.ConvertUint64ToString(r.UserIds)
	event.PublishChs = []string{strconv.Itoa(int(r.RoomIds[0]))}
	event.Kind = true
	return err
}
func DeleteUsersFromRoom(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r, err := models.UnmarshalEvent[models.RoomRequest](event.Data)
	if err != nil {
		return err
	}
	if len(r.RoomIds) == 0 || len(r.UserIds) == 0 {
		return errors.New("malformed json")
	}
	if len(r.UserIds) != 1 || r.UserIds[0] != event.UserId { // if the user wants to remove themselves from the room skip this check
		if err := tx.QueryRow(ctx, `SELECT 1 FROM rooms WHERE room_id = $1 AND created_by_user_id = $2 FOR UPDATE`,
			r.RoomIds[0], event.UserId).Scan(new(int)); err != nil {
			err = errors.New("you have no permission to delete users from this room")
			if err != nil {
				return err
			}
		}
	}
	tag, err := tx.Exec(ctx, deleteUsersFromRoom, r.UserIds, r.RoomIds[0], true) // bool param is is_group
	if err != nil {
		return err
	}
	if int(tag.RowsAffected()) != len(r.UserIds) {
		return errors.New("no users added")
	}
	rows, err := tx.Query(ctx, getRoomUsers, r.RoomIds[0]) // bool param is is_group
	if err != nil {
		return err
	}

	resp, err := NormalizeRoom(rows, r.RoomIds[0])
	if err != nil {
		return err
	}
	if len(resp) == 0 {
		return errors.New("no users deleted")
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 1
	event.OrderCmd[1] = 2
	event.UserChs = models.ConvertUint64ToString(r.UserIds)
	event.PublishChs = []string{strconv.Itoa(int(r.RoomIds[0]))}
	event.Kind = false
	return err
}

// Blocking user and delete user from room_users_info
func BlockUser(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r, err := models.UnmarshalEvent[models.RoomRequest](event.Data)
	if err != nil {
		return err
	}
	r.RoomIds = make([]uint64, 1)
	if len(r.UserIds) == 0 {
		return fmt.Errorf("malformed json")
	}
	first, second := r.UserIds[0], event.UserId
	if event.UserId > r.UserIds[0] {
		first, second = event.UserId, r.UserIds[0]
	}
	r.UserIds = append(r.UserIds, event.UserId)
	userstr := models.ConvertUint64ToString(r.UserIds)

	row := tx.QueryRow(ctx, `SELECT room_id
		FROM duo_rooms
		WHERE (user_id1 = $1 AND user_id2 = $2) OR (user_id2 = $1 AND user_id1 = $2) FOR UPDATE`, first, second)

	err = row.Scan(&r.RoomIds[0])
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	tag, err := tx.Exec(ctx, deleteUsersFromRoom, r.UserIds, r.RoomIds[0], false) // bool param is is_group
	if int(tag.RowsAffected()) != len(r.UserIds) || err != nil {
		return errors.New("no users added: " + err.Error())
	}
	t, err := tx.Exec(ctx, `INSERT INTO blocked_users (blocked_by_user_id, blocked_user_id)
		VALUES ($1, $2)`, event.UserId, r.UserIds[0])
	if err != nil {
		return err
	}
	if t.RowsAffected() == 0 {
		return errors.New("no users blocked")
	}
	rows, err := tx.Query(ctx, getRoomUsers, r.RoomIds[0])
	if err != nil {
		return err
	}
	resp, err := NormalizeRoom(rows, r.RoomIds[0])
	if err != nil {
		return err
	}
	if len(resp) == 0 {
		return errors.New("no users blocked")
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 1
	event.OrderCmd[1] = 2
	event.UserChs = userstr
	event.PublishChs = []string{strconv.Itoa(int(r.RoomIds[0]))}
	event.Kind = false

	return err
}

// Unblocking user
func UnblockUser(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r, err := models.UnmarshalEvent[models.RoomRequest](event.Data)
	if err != nil {
		return err
	}
	if len(r.UserIds) == 0 {
		return errors.New("malformed json")
	}
	tag, err := tx.Exec(ctx, `DELETE FROM blocked_users 
			WHERE blocked_by_user_id = $1 AND blocked_user_id = $2`, event.UserId, r.UserIds[0])
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("no blocked users found")
	}
	event.Type = 4
	return err
}
func ChangeRoomname(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) (err error) {
	r, err := models.UnmarshalEvent[models.RoomRequest](event.Data)
	if err != nil {
		return err
	}
	if len(r.RoomIds) == 0 || len(r.RoomName) == 0 {
		return errors.New("malformed json")
	}
	r.RoomName, err = NormalizeString(r.RoomName)
	if err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `UPDATE rooms SET room_name = $1 WHERE room_id = $2 AND created_by_user_id = $3`, r.RoomName, r.RoomIds[0], event.UserId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("'change room name' hasn't changed")
	}
	rm := []models.RoomClient{{RoomName: r.RoomName, RoomId: r.RoomIds[0]}}
	event.Data, err = json.Marshal(rm)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 1
	event.PublishChs = []string{strconv.Itoa(int(r.RoomIds[0]))}
	return err
}

func NormalizeRoom(rows pgx.Rows, roomIdOpt uint64) ([]models.RoomClient, error) {
	var rms []models.RoomClient
	var err error
	var currentroom uint64
	defer rows.Close()
	for rows.Next() {
		r, err := pgx.RowToStructByNameLax[models.RoomClient](rows)
		if err != nil {
			return nil, err
		}
		if r.RoomId == currentroom {
			last := len(rms) - 1
			rms[last].Users = append(rms[last].Users, models.User{UserId: r.UserId, Username: r.Username})
		} else {
			rms = append(rms, r)
			last := len(rms) - 1
			rms[last].Users = make([]models.User, 0, 3)
			rms[last].Users = append(rms[last].Users, models.User{UserId: r.UserId, Username: r.Username})
			currentroom = r.RoomId
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(rms) == 0 && roomIdOpt != 0 { // if room exist but room have no users
		rms = append(rms, models.RoomClient{RoomId: roomIdOpt})
	}
	return rms, err
}
