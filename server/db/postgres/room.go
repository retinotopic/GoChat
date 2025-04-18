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

type RoomRequest struct {
	UserIds  []uint64 `json:"UserIds" `
	RoomIds  []uint64 `json:"RoomIds" `
	RoomName string   `json:"RoomName" `
	IsGroup  bool     `json:"IsGroup" `
}
type RoomClient struct {
	RoomId          uint64 `json:"RoomId" `
	RoomName        string `json:"RoomName" `
	IsGroup         bool   `json:"IsGroup" `
	CreatedByUserId uint64 `json:"CreatedByUserId" `
	Users           []User `json:"Users" `
	Username        string `json:"-" `
	UserId          uint64 `json:"-" `
}

func GetAllRooms(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) (err error) {
	rows, err := tx.Query(ctx,
		`SELECT ru.room_id,ru.user_id,r.room_name,r.is_group,r.created_by_user_id,u.user_name
		FROM room_users_info ru JOIN rooms r ON ru.room_id = r.room_id JOIN users u ON ru.user_id = u.user_id
		WHERE ru.user_id = $1 
		ORDER BY r.last_activity DESC, r.room_id;
		`, event.UserId)
	if err != nil {
		return err
	}
	resp, err := NormalizeRoom(rows, false)
	if err != nil {
		return err
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 2
	event.PubForSub = []string{strconv.Itoa(int(event.UserId))} //1
	for _, room := range resp {
		event.SubForPub = append(event.SubForPub, "room"+strconv.Itoa(int(room.RoomId))) //2
	}
	event.Kind = "1"
	return err
}
func (r *RoomRequest) IsDuoRoomExist(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	first, second := r.UserIds[0], event.UserId
	if event.UserId > r.RoomIds[0] {
		first, second = event.UserId, r.UserIds[0]
	}
	row := tx.QueryRow(ctx, `SELECT room_id
		FROM duo_rooms
		WHERE (user_id1 = $1 AND user_id2 = $2) OR (user_id2 = $1 AND user_id1 = $2) ;`, first, second)
	err := row.Scan(r.RoomIds[0])
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	return nil
}

// method for safely creating unique duo room
func CreateDuoRoom(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r := &RoomRequest{}

	err := json.Unmarshal(event.Data, r)
	if err != nil {
		return err
	}
	r.RoomIds = make([]uint64, 1)
	if len(r.UserIds) == 0 {
		return errors.New("malformed json")
	}
	err = r.IsDuoRoomExist(ctx, tx, event)
	if err != nil {
		return err
	}
	var rows pgx.Rows
	is_group := false
	if r.RoomIds[0] == 0 {
		err := tx.QueryRow(ctx, CreateRoom, r.RoomName, is_group, event.UserId).Scan(&r.RoomIds[0])
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO duo_rooms (user_id1,user_id2,room_id) VALUES ($1,$2,$3)`, event.UserId, r.UserIds[0], r.RoomIds[0])
		if err != nil {
			return err
		}
	}
	rows, err = tx.Query(ctx, addUsersToRoomDirect, r.RoomIds[0], r.UserIds, true, event.UserId) // bool param is is_group
	if err != nil {
		return err
	}

	resp, err := NormalizeRoom(rows, false)
	if err != nil {
		return err
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 2
	event.OrderCmd[1] = 1
	event.PubForSub = ConvertUint64ToString(r.UserIds)
	event.SubForPub = []string{"room" + strconv.Itoa(int(r.RoomIds[0]))}
	event.Kind = "1"
	return err
}

func CreateGroupRoom(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r := &RoomRequest{}
	err := json.Unmarshal(event.Data, r)
	if err != nil {
		return err
	}
	r.RoomIds = make([]uint64, 1)
	if len(r.RoomName) == 0 || len(r.UserIds) == 0 {
		return errors.New("malformed json")
	}
	// create new room and return room id
	err = tx.QueryRow(ctx, CreateRoom, r.RoomName, true, event.UserId).Scan(r.RoomIds[0]) // bool param is is_group
	if err != nil {
		return err
	}
	rows, err := tx.Query(ctx, addUsersToRoomGroup, r.RoomIds[0], r.UserIds, true, event.UserId) // bool param is is_group
	if err != nil {
		return err
	}
	resp, err := NormalizeRoom(rows, false)
	if err != nil {
		return err
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 2
	event.OrderCmd[1] = 1
	event.PubForSub = ConvertUint64ToString(r.UserIds)
	event.SubForPub = []string{"room" + strconv.Itoa(int(r.RoomIds[0]))}
	event.Kind = "1"
	return err
}
func AddUsersToRoom(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r := &RoomRequest{}
	err := json.Unmarshal(event.Data, r)
	if err != nil {
		return err
	}
	if len(r.RoomIds) == 0 || len(r.UserIds) == 0 {
		return errors.New("malformed json")
	}
	if err := tx.QueryRow(ctx, `SELECT 1 FROM rooms WHERE room_id = $1 AND created_by_user_id = $2`, r.RoomIds[0], event.UserId).Scan(new(int)); err != nil {
		err = errors.New("you have no permission to add users to this room")
		return err
	}
	rows, err := tx.Query(ctx, addUsersToRoomGroup, r.RoomIds[0], r.UserIds, true, event.UserId) // bool param is is_group
	if err != nil {
		return err
	}
	resp, err := NormalizeRoom(rows, false)
	if err != nil {
		return err
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 2
	event.OrderCmd[1] = 1
	event.PubForSub = ConvertUint64ToString(r.UserIds)
	event.SubForPub = []string{"room" + strconv.Itoa(int(r.RoomIds[0]))}
	event.Kind = "1"
	return err
}
func DeleteUsersFromRoom(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r := &RoomRequest{}
	err := json.Unmarshal(event.Data, r)
	if err != nil {
		return err
	}
	if len(r.RoomIds) == 0 || len(r.UserIds) == 0 {
		return errors.New("malformed json")
	}
	if len(r.UserIds) != 1 || r.UserIds[0] != event.UserId { // if the user wants to remove themselves from the room skip this check
		if err := tx.QueryRow(ctx, `SELECT 1 FROM rooms WHERE room_id = $1 AND created_by_user_id = $2`, r.RoomIds[0], event.UserId).Scan(new(int)); err != nil {
			err = errors.New("you have no permission to delete users from this room")
			if err != nil {
				return err
			}
		}
	}
	rows, err := tx.Query(ctx, deleteUsersFromRoom, r.UserIds, r.RoomIds[0], true) // bool param is is_group
	if err != nil {
		return err
	}

	resp, err := NormalizeRoom(rows, true)
	if err != nil {
		return err
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 1
	event.OrderCmd[1] = 2
	event.PubForSub = ConvertUint64ToString(r.UserIds)
	event.SubForPub = []string{"room" + strconv.Itoa(int(r.RoomIds[0]))}
	event.Kind = "0"
	return err
}

// Blocking user and delete user from room_users_info
func BlockUser(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r := &RoomRequest{}
	err := json.Unmarshal(event.Data, r)
	if err != nil {
		return err
	}
	r.RoomIds = make([]uint64, 1)
	if len(r.UserIds) == 0 {
		return fmt.Errorf("malformed json")
	}

	_, err = tx.Exec(ctx, deleteUsersFromRoom, r.UserIds, r.RoomIds[0], false) // bool param is is_group
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 1
	event.OrderCmd[1] = 2
	event.PubForSub = ConvertUint64ToString(r.UserIds)
	event.SubForPub = []string{"room" + strconv.Itoa(int(r.RoomIds[0]))}
	event.Kind = "0"

	_, err = tx.Exec(ctx, `INSERT INTO blocked_users (blocked_by_user_id, blocked_user_id)
		VALUES ($1, $2)`, event.UserId, r.UserIds[0])
	if err != nil {
		return err
	}
	return err
}

// Unblocking user
func UnblockUser(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r := &RoomRequest{}
	err := json.Unmarshal(event.Data, r)
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
		return errors.New("internal database error, no blocked users found")
	}
	return err
}
func ChangeRoomname(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	r := &RoomRequest{}
	err := json.Unmarshal(event.Data, r)
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
		return errors.New("internal database error, 'change room name' hasn't changed")
	}
	event.Data, err = json.Marshal(r)
	if err != nil {
		return err
	}
	event.OrderCmd[0] = 1
	event.SubForPub = []string{"room" + strconv.Itoa(int(r.RoomIds[0]))}
	return err
}
func ConvertUint64ToString(ids []uint64) []string {
	if len(ids) == 0 {
		return []string{}
	}

	strIds := make([]string, len(ids)+1)
	for i, id := range ids {
		strIds[i] = strconv.FormatUint(uint64(id), 10)
	}
	return strIds
}
func NormalizeRoom(rows pgx.Rows, userDelete bool) ([]RoomClient, error) {
	var rms []RoomClient
	var err error
	var currentroom uint64
	defer rows.Close()
	for rows.Next() {
		r, err := pgx.RowToStructByNameLax[RoomClient](rows)
		if err != nil {
			return nil, err
		}

		if r.RoomId == currentroom {
			last := len(rms) - 1
			rms[last].Users = append(rms[last].Users, User{UserId: r.UserId, Username: r.Username, RoomToggle: userDelete})
		} else {
			rms = append(rms, r)
			last := len(rms) - 1
			rms[last].Users = make([]User, 0, 3)
			rms[last].Users = append(rms[last].Users, User{UserId: r.UserId, Username: r.Username, RoomToggle: userDelete})
			currentroom = r.RoomId
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rms, err
}
