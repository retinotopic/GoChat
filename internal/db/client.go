package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (p *PgClient) GetAllRooms(ctx context.Context, flowjson *FlowJSON) (err error) {
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
	Rows, err := p.Query(ctx,
		`SELECT r.room_id,r.name
		FROM room_users_info ru JOIN rooms r ON ru.room_id = r.room_id
		WHERE ru.user_id = $1 
		ORDER BY r.last_activity DESC;
		`, p.UserID)
	if err != nil {
		return err
	}

	err = Rows.Err()
	if err != nil {
		return err
	}

	for Rows.Next() {
		err := Rows.Scan(&flowjson.Room)
		if err != nil {
			return err
		}
		p.Chan <- *flowjson
		p.RoomsPagination = append(p.RoomsPagination, flowjson.Room)
	}

	return err
}

// load messages from a room
func (p *PgClient) GetMessagesFromRoom(ctx context.Context, flowjson *FlowJSON) error {
	var Rows pgx.Rows
	Rows, err := p.Query(ctx,
		`SELECT payload,user_id,
		FROM messages 
		WHERE room_id = $1 AND message_id < $2
		ORDER BY message_id DESC`, flowjson.Room, flowjson.MessageId)
	if err != nil {
		return err
	}
	p.toChannel(flowjson, Rows, flowjson.Message, flowjson.Users[0])
	return err
}

func (p *PgClient) GetNextRooms(ctx context.Context, flowjson *FlowJSON) error {
	p.NRMutex.Lock()
	defer p.NRMutex.Unlock()
	var arrayrooms []uint32
	for i := p.PaginationOffset; i < p.PaginationOffset+30; i++ {
		arrayrooms = append(arrayrooms, p.RoomsPagination[i])
	}
	var Rows pgx.Rows
	Rows, err := p.Query(ctx,
		`SELECT room_id,name FROM rooms WHERE room_id IN ($1)`, arrayrooms)
	if err != nil {
		return err
	}
	//pgx.CollectOneRow(Rows, pgx.RowToAddrOfStructByName[FlowJSON])
	p.toChannel(flowjson, Rows, flowjson.Room, flowjson.Name)

	p.PaginationOffset += 30
	return err
}

func (p *PgClient) GetRoomUsersInfo(ctx context.Context, flowjson *FlowJSON) error {
	var Rows pgx.Rows
	Rows, err := p.Query(ctx,
		`SELECT u.user_id,u.name
		FROM users u JOIN room_users_info ru ON ru.user_id = u.user_id
		WHERE ru.room_id = $1`, p.UserID)
	if err != nil {
		return err
	}
	p.toChannel(flowjson, Rows, &flowjson.User, &flowjson.Name)
	return err
}

func (p *PgClient) FindUsers(ctx context.Context, flowjson *FlowJSON) error {
	var Rows pgx.Rows
	Rows, err := p.Query(ctx,
		`SELECT user_id,username FROM users WHERE username ILIKE $1 LIMIT 20`, flowjson.Name+"%")
	if err != nil {
		return err
	}
	p.toChannel(flowjson, Rows, &flowjson.User, &flowjson.Name)
	return err
}
func (c *PgClient) toChannel(flowjson *FlowJSON, rows pgx.Rows, dest ...any) {
	err := rows.Err() // checking for query timeout
	if err != nil {
		flowjson.ErrorMsg = err.Error()
		return
	}
	for rows.Next() {
		err := rows.Scan(dest...)
		if err != nil {
			flowjson.ErrorMsg = err.Error()
			return
		}
		c.Chan <- *flowjson
	}
}
