package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func (p *PgClient) GetAllRooms(ctx context.Context, flowjson *FlowJSON) {
	p.Once.Do(func() {
		var Rows pgx.Rows
		Rows, flowjson.Err = p.Query(ctx,
			`SELECT r.room_id,r.name
			FROM room_users_info ru JOIN rooms r ON ru.room_id = r.room_id
			WHERE ru.user_id = $1 
			ORDER BY r.last_activity DESC;
			`, p.UserID)
		err := Rows.Err()
		if err != nil {
			log.Println("query timeout", flowjson.Err)
			flowjson.Err = err
			return
		}

		for Rows.Next() {
			err := Rows.Scan(&flowjson.Room)
			if err != nil {
				log.Println("Error scanning rows:", err)
				flowjson.Err = err
				return
			}
			p.Chan <- flowjson
			p.RoomsPagination = append(p.RoomsPagination, flowjson.Room)
		}
	})

}

// load messages from a room
func (p *PgClient) GetMessagesFromRoom(ctx context.Context, flowjson *FlowJSON) {
	var Rows pgx.Rows
	Rows, flowjson.Err = p.Query(ctx,
		`SELECT payload,user_id,
		FROM messages 
		WHERE room_id = $1 AND message_id < $2
		ORDER BY message_id DESC`, flowjson.Room, flowjson.Message_id)
	if flowjson.Err != nil {
		log.Println("Error getting messages from this rooms:", flowjson.Err)
		return
	}
	p.toChannel(flowjson, Rows, flowjson.Message, flowjson.Users[0])
}

func (p *PgClient) GetNextRooms(ctx context.Context, flowjson *FlowJSON) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	var arrayrooms []uint32
	for i := p.PaginationOffset; i < p.PaginationOffset+30; i++ {
		arrayrooms = append(arrayrooms, p.RoomsPagination[i])
	}
	var Rows pgx.Rows
	Rows, flowjson.Err = p.Query(ctx,
		`SELECT room_id,name FROM rooms WHERE room_id IN ($1)`, arrayrooms)
	if flowjson.Err != nil {
		log.Println("Error getting messages from this rooms:", flowjson.Err)
		return
	}
	pgx.CollectOneRow(Rows, pgx.RowToAddrOfStructByName[FlowJSON])
	p.toChannel(flowjson, Rows, flowjson.Room, flowjson.Name)
	if flowjson.Err == nil {
		p.PaginationOffset += 30
	}
}

func (p *PgClient) GetRoomUsersInfo(ctx context.Context, flowjson *FlowJSON) {
	var Rows pgx.Rows
	Rows, flowjson.Err = p.Query(ctx,
		`SELECT u.user_id,u.name
		FROM users u JOIN room_users_info ru ON ru.user_id = u.user_id
		WHERE ru.room_id = $1`, p.UserID)
	if flowjson.Err != nil {
		log.Println("Error getting room info", flowjson.Err)
		return
	}
	p.toChannel(flowjson, Rows, &flowjson.User, &flowjson.Name)
}

func (p *PgClient) FindUsers(ctx context.Context, flowjson *FlowJSON) {
	var Rows pgx.Rows
	Rows, flowjson.Err = p.Query(ctx,
		`SELECT user_id,username FROM users WHERE username ILIKE $1 LIMIT 20`, flowjson.Name+"%")
	p.toChannel(flowjson, Rows, &flowjson.User, &flowjson.Name)
}
func (c *PgClient) toChannel(flowjson *FlowJSON, rows pgx.Rows, dest ...any) {
	err := rows.Err() // checking for query timeout
	if err != nil {
		flowjson.Err = err
		return
	}
	for rows.Next() {
		err := rows.Scan(dest...)
		if err != nil {
			log.Println("Error scanning rows:", err)
			flowjson.Err = err
			return
		}
		c.Chan <- flowjson
	}
}
