package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/goccy/go-json"

	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/internal/models"
)

type RoomRequest struct {
	UserIds []uint32 `json:"UserIds" `
	RoomIds []uint32 `json:"RoomIds" `
	Name    string   `json:"Name" `
	IsGroup bool     `json:"IsGroup" `
}
type RoomClient struct {
	RoomId   uint32 `json:"RoomId" `
	UserId   uint32 `json:"UserId" `
	Name     string `json:"Name" `
	IsGroup  bool   `json:"IsGroup" `
	Username string `json:"Username" `
}

// uint32[] rooms
func GetNextRooms(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	r := &RoomRequest{}
	err := json.Unmarshal(event.Data, r)
	if err != nil {
		return err
	}
	if len(r.RoomIds) == 0 {
		return fmt.Errorf("malformed json")
	}
	rows, err := tx.Query(ctx,
		`SELECT ru.room_id,ru.user_id,r.name,r.is_group,u.username FROM room_users_info ru JOIN rooms r ON ru.room_id = r.room_id JOIN users u ON ru.user_id = u.user_id
WHERE ru.room_id = ANY($1) ORDER BY ru.room_id`, r.RoomIds)
	if err != nil {
		return err
	}
	resp, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[RoomClient])
	if err != nil {
		return err
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	return err
}

func GetAllRoomsIds(ctx context.Context, tx pgx.Tx, event *models.Event) (err error) {
	rows, err := tx.Query(ctx,
		`SELECT r.room_id
		FROM room_users_info ru JOIN rooms r ON ru.room_id = r.room_id
		WHERE ru.user_id = $1 
		ORDER BY r.last_activity DESC;
		`, event.UserId)

	resp, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[RoomClient])
	if err != nil {
		return err
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	return err
}

// method for safely creating unique duo room
func CreateDuoRoom(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	r := &RoomRequest{}

	err := json.Unmarshal(event.Data, r)
	if err != nil {
		return err
	}
	r.RoomIds = make([]uint32, 1)
	if len(r.UserIds) == 0 {
		return fmt.Errorf("malformed json")
	}
	err = r.IsDuoRoomExist(ctx, tx, event)
	if err != nil {
		return err
	}

	is_group := false
	if r.RoomIds[0] == 0 {
		err := tx.QueryRow(ctx, CreateRoom, r.Name, is_group, event.UserId).Scan(&r.RoomIds[0])
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO duo_rooms (user_id1,user_id2,room_id) VALUES ($1,$2,$3)`, event.UserId, r.UserIds[0], r.RoomIds[0])
		if err != nil {
			return err
		}
	} else {
		addUsersToRoom := fmt.Sprintf(addUsersToRoom, allowDirectMessages)
		_, err := tx.Exec(ctx, addUsersToRoom, r.RoomIds[0], r.UserIds, true, event.UserId) // bool param is is_group
		if err != nil {
			return err
		}
	}
	return err
}
func (r *RoomRequest) IsDuoRoomExist(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	row := tx.QueryRow(ctx, `SELECT room_id
		FROM duo_rooms
		WHERE (user_id1 = $1 AND user_id2 = $2) OR (user_id2 = $1 AND user_id1 = $2) ;`, event.UserId, r.UserIds[0])
	err := row.Scan(r.RoomIds[0])
	if err != nil && err != pgx.ErrNoRows {
		return err
	}
	return nil
}
func CreateGroupRoom(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	r := &RoomRequest{}
	err := json.Unmarshal(event.Data, r)
	if err != nil {
		return err
	}
	r.RoomIds = make([]uint32, 1)
	if len(r.Name) == 0 || len(r.UserIds) == 0 {
		return fmt.Errorf("malformed json")
	}
	// create new room and return room id
	err = tx.QueryRow(ctx, CreateRoom, r.Name, true, event.UserId).Scan(r.RoomIds[0]) // bool param is is_group
	if err != nil {
		return err
	}
	addUsersToRoom := fmt.Sprintf(addUsersToRoom, allowGroupInvites)
	_, err = tx.Exec(ctx, addUsersToRoom, r.RoomIds[0], r.UserIds, true, event.UserId) // bool param is is_group
	if err != nil {
		return err
	}
	return err
}
func AddUsersToRoom(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	r := &RoomRequest{}
	err := json.Unmarshal(event.Data, r)
	if err != nil {
		return err
	}
	if len(r.RoomIds) == 0 || len(r.UserIds) == 0 {
		return fmt.Errorf("malformed json")
	}
	if err := tx.QueryRow(ctx, `SELECT 1 FROM rooms WHERE room_id = $1 AND created_by_user_id = $2`, r.RoomIds[0], event.UserId).Scan(new(int)); err != nil {
		err = errors.New("you have no permission to add users to this room")
		return err
	}
	addUsersToRoom := fmt.Sprintf(addUsersToRoom, allowGroupInvites)
	_, err = tx.Exec(ctx, addUsersToRoom, r.RoomIds[0], r.UserIds, true, event.UserId) // bool param is is_group
	if err != nil {
		return err
	}
	return err
}
func DeleteUsersFromRoom(ctx context.Context, tx pgx.Tx, event models.Event) error {
	r := &RoomRequest{}
	err := json.Unmarshal(event.Data, r)
	if err != nil {
		return err
	}
	if len(r.RoomIds) == 0 || len(r.UserIds) == 0 {
		return fmt.Errorf("malformed json")
	}
	if len(r.UserIds) != 1 || r.UserIds[0] != event.UserId { // first if statement, if the user wants to remove themselves from the room
		if err := tx.QueryRow(ctx, `SELECT 1 FROM rooms WHERE room_id = $1 AND created_by_user_id = $2`, r.RoomIds[0], event.UserId).Scan(new(int)); err != nil {
			err = errors.New("you have no permission to delete users from this room")
			if err != nil {
				return err
			}
		}
	}
	_, err = tx.Exec(ctx, deleteUsersFromRoom, r.UserIds, r.RoomIds[0], true) // bool param is is_group
	if err != nil {
		return err
	}
	return err
}

// Blocking user and delete user from duo room
func (p *PgClient) BlockUser(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	r := &RoomRequest{}
	err := json.Unmarshal(event.Data, r)
	if err != nil {
		return err
	}
	r.RoomIds = make([]uint32, 1)
	if len(r.UserIds) == 0 {
		return fmt.Errorf("malformed json")
	}
	err = r.IsDuoRoomExist(ctx, tx, event)
	if err != nil {
		return err
	}
	if r.RoomIds[0] != 0 {
		_, err = tx.Exec(ctx, deleteUsersFromRoom, r.UserIds, r.RoomIds[0], false) // bool param is is_group
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO blocked_users (blocked_by_user_id, blocked_user_id)
		VALUES ($1, $2)`, event.UserId, r.UserIds[0])
	if err != nil {
		return err
	}
	return err
}

// Unblocking user
func (c *PgClient) UnblockUser(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	r := &RoomRequest{}
	err := json.Unmarshal(event.Data, r)
	if err != nil {
		return err
	}
	if len(r.UserIds) == 0 {
		return fmt.Errorf("malformed json")
	}
	_, err = tx.Exec(ctx, `DELETE FROM blocked_users 
			WHERE blocked_by_user_id = $1 AND blocked_user_id = $2`, event.UserId, r.UserIds[0])
	if err != nil {
		return err
	}
	return err
}
