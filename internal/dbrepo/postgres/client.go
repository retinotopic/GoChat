package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/internal/models"
)

func (p *PgClient) GetAllRooms(ctx context.Context, flowjson *models.Flowjson) (err error) {
	p.GRMutex.Lock()
	defer func() {
		if err != nil {
			p.ReOnce = false
		}
		p.GRMutex.Unlock()
	}()
	if p.ReOnce {
		return err
	}
	rows, err := p.Query(ctx,
		`SELECT r.room_id,r.name
		FROM room_users_info ru JOIN rooms r ON ru.room_id = r.room_id
		WHERE ru.user_id = $1 
		ORDER BY r.last_activity DESC;
		`, p.UserID)

	fjarr, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.Flowjson])
	if err != nil {
		return err
	}
	for _, v := range fjarr {
		p.Chan <- v
		p.RoomsPagination = append(p.RoomsPagination, flowjson.Room)
	}

	return err
}

// load messages from a room
func (p *PgClient) GetMessagesFromRoom(ctx context.Context, flowjson *models.Flowjson) error {
	rows, err := p.Query(ctx,
		`SELECT payload,user_id,
		FROM messages 
		WHERE room_id = $1 AND message_id < $2
		ORDER BY message_id DESC`, flowjson.Room, flowjson.MessageId)
	if err != nil {
		return err
	}
	err = p.toChannel(rows)
	if err != nil {
		return err
	}
	return err
}

func (p *PgClient) GetNextRooms(ctx context.Context, flowjson *models.Flowjson) error {
	p.NRMutex.Lock()
	defer p.NRMutex.Unlock()
	var arrayrooms []uint32
	for i := p.PaginationOffset; i < p.PaginationOffset+30; i++ {
		arrayrooms = append(arrayrooms, p.RoomsPagination[i])
	}

	rows, err := p.Query(ctx,
		`SELECT room_id,name FROM rooms WHERE room_id IN ($1)`, arrayrooms)
	if err != nil {
		return err
	}
	err = p.toChannel(rows)
	if err != nil {
		return err
	}

	p.PaginationOffset += 30
	return err
}

func (p *PgClient) GetRoomUsersInfo(ctx context.Context, flowjson *models.Flowjson) error {
	rows, err := p.Query(ctx,
		`SELECT u.user_id,u.name
		FROM users u JOIN room_users_info ru ON ru.user_id = u.user_id
		WHERE ru.room_id = $1`, p.UserID)
	if err != nil {
		return err
	}
	err = p.toChannel(rows)
	if err != nil {
		return err
	}
	return err
}

func (p *PgClient) FindUsers(ctx context.Context, flowjson *models.Flowjson) error {
	rows, err := p.Query(ctx,
		`SELECT user_id,username FROM users WHERE username ILIKE $1 LIMIT 20`, flowjson.Name+"%")
	if err != nil {
		return err
	}
	err = p.toChannel(rows)
	if err != nil {
		return err
	}
	return err
}
func (p *PgClient) toChannel(rows pgx.Rows) error {
	fjarr, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.Flowjson])
	if err != nil {
		return err
	}
	for _, v := range fjarr {
		p.Chan <- v
	}
	return err
}
