package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/internal/models"
)

// method for safely creating unique duo room
func (p *PgClient) CreateDuoRoom(ctx context.Context, tx pgx.Tx, fj *models.Flowjson) error {
	err := p.IsDuoRoomExist(ctx, tx, fj)
	if err != nil {
		return err
	}
	if fj.RoomId == 0 {
		err = p.CreateRoom(ctx, tx, fj)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO duo_rooms (user_id1,user_id2,room_id) VALUES ($1,$2,$3)`, fj.Users[0], fj.Users[1], fj.RoomId)
	} else {
		p.AddUsersToRoom(ctx, tx, fj)
	}
	return err
}
func (p *PgClient) IsDuoRoomExist(ctx context.Context, tx pgx.Tx, fj *models.Flowjson) error {
	var rows pgx.Rows
	rows, err := tx.Query(ctx, `SELECT room_id
		FROM duo_rooms
		WHERE (user_id1 = $1 AND user_id2 = $2) OR (user_id2 = $1 AND user_id1 = $2) ;`, fj.Users[0], fj.Users[1])
	if err != nil {
		return err
	}
	if rows.Next() {
		err = rows.Scan(&fj.RoomId)
		defer rows.Close()
		if err != nil {
			return err
		}
	}
	return err
}
func (p *PgClient) CreateRoom(ctx context.Context, tx pgx.Tx, fj *models.Flowjson) error {
	var is_group bool
	if fj.Mode != "CreateDuoRoom" {
		is_group = true
	}
	// create new room and return room id
	err := tx.QueryRow(ctx, "INSERT INTO rooms (name,is_group,created_by_user_id) VALUES ($1,$2,$3) RETURNING room_id", fj.Name, is_group, p.UserID).Scan(&fj.RoomId)
	if err != nil {
		return err
	}
	err = p.AddUsersToRoom(ctx, tx, fj)
	if err != nil {
		return err
	}
	return err
}
func (p *PgClient) AddUsersToRoom(ctx context.Context, tx pgx.Tx, fj *models.Flowjson) error {
	var rows pgx.Rows
	if fj.Mode == "AddUsersToRoom" {
		if err := tx.QueryRow(ctx, `SELECT 1 FROM rooms WHERE room_id = $1 AND created_by_user_id = $2`, fj.RoomId, p.UserID).Scan(new(int)); err != nil {
			err = errors.New("you have no permission to add users to this room")
			if err != nil {
				return err
			}
		}
	}
	query := `INSERT INTO room_users_info (room_id,user_id)
			SELECT $1,users_to_add.user_id
			FROM (SELECT unnest($2::int[]) AS user_id) AS users_to_add
			JOIN users u ON u.user_id = users_to_add.user_id AND %s
			JOIN rooms r ON r.room_id = $1 AND r.is_group = $3
			LEFT JOIN blocked_users bu ON (bu.blocked_by_user_id = users_to_add.user_id AND bu.blocked_user_id = $4) 
			OR (bu.blocked_by_user_id = $4 AND bu.blocked_user_id = users_to_add.user_id )
			WHERE bu.blocked_by_user_id IS NULL RETURNING user_id;`

	var condition string
	var is_group bool
	if fj.Mode == "CreateDuoRoom" {
		condition = "u.allow_direct_messages = true"
	} else {
		condition = "u.allow_group_invites = true"
		is_group = true
	}
	query = fmt.Sprintf(query, condition)
	rows, err := tx.Query(ctx, query, fj.RoomId, fj.Users, is_group, p.UserID)
	if err != nil {
		return err
	}
	counter := 0
	for rows.Next() {
		counter++
	}
	if len(fj.Users) != counter {
		err = errors.New("at least one user cannot be added to the room")
	}
	return err
}
func (p *PgClient) DeleteUsersFromRoom(ctx context.Context, tx pgx.Tx, fj *models.Flowjson) error {
	if fj.Mode == "DeleteUsersFromRoom" {
		if len(fj.Users) != 1 || fj.Users[0] != p.UserID {
			if err := tx.QueryRow(ctx, `SELECT 1 FROM rooms WHERE room_id = $1 AND created_by_user_id = $2`, fj.RoomId, p.UserID).Scan(new(int)); err != nil {
				err = errors.New("you have no permission to delete users from this room")
				if err != nil {
					return err
				}
			}
		}
	}
	query := `DELETE FROM room_users_info
	WHERE user_id = ANY($1)
	AND room_id IN (
		SELECT room_id
		FROM rooms 
		WHERE room_id = $2 AND is_group = $3
	);`
	var is_group bool

	if fj.Mode != "BlockUser" {
		is_group = true
	}
	_, err := tx.Exec(ctx, query, fj.Users, fj.RoomId, is_group)
	if err != nil {
		return err
	}
	return err
}

// Blocking user and delete user from duo room
func (p *PgClient) BlockUser(ctx context.Context, tx pgx.Tx, fj *models.Flowjson) error {

	err := p.IsDuoRoomExist(ctx, tx, fj)

	if err != nil {
		return err
	}
	if fj.RoomId != 0 {
		err := p.DeleteUsersFromRoom(ctx, tx, fj)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `INSERT INTO blocked_users (blocked_by_user_id, blocked_user_id)
		VALUES ($1, $2)`, fj.Users[0], fj.Users[1])
	if err != nil {
		return err
	}
	return err
}

// Unblocking user
func (c *PgClient) UnblockUser(ctx context.Context, tx pgx.Tx, fj *models.Flowjson) error {
	_, err := tx.Exec(ctx, `DELETE FROM blocked_users 
			WHERE blocked_by_user_id = $1 AND blocked_user_id = $2`, c.UserID, fj.Users[1])
	if err != nil {
		return err
	}
	return err
}
