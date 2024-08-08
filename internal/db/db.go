package db

import (
	"context"
	"log"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/retinotopic/GoChat/pkg/str"
)

type FlowJSON struct {
	Mode       string   `json:"Mode"`
	Message    string   `json:"Message"`
	Users      []uint32 `json:"Users"`
	Room       uint32   `json:"Room"`
	Name       string   `json:"Name"`
	Message_id string   `json:"Offset"`
	ErrorMsg   string   `json:"ErrorMsg"`
	Bool       bool     `json:"Bool"`
	Tx         pgx.Tx
	Err        error
}

type PgClient struct {
	Sub              string
	Name             string
	UserID           uint32
	Mutex            sync.Mutex
	RoomsPagination  []uint32
	RoomsCount       uint8 // no more than 250
	PaginationOffset uint8
	funcmap          map[string]fnAPI
	Chan             chan *FlowJSON
	*pgxpool.Pool
}
type fnAPI struct {
	Fn         func(context.Context, *FlowJSON)
	ClientOnly bool
}

func (p *PgClient) SendMessage(ctx context.Context, flowjson *FlowJSON) {
	_, flowjson.Err = p.Exec(ctx, `INSERT INTO messages (payload,user_id,room_id) VALUES ($1,$2,$3)`, flowjson.Message, p.UserID, flowjson.Room)
	if flowjson.Err != nil {
		log.Println("Error inserting message:", flowjson.Err)
		return
	}
}
func (p *PgClient) ChangeUsername(ctx context.Context, flowjson *FlowJSON) {
	username := str.NormalizeString(flowjson.Name)
	_, flowjson.Err = p.Exec(ctx, "UPDATE users SET username = $1 WHERE user_id = $2", username, p.UserID)
	if flowjson.Err != nil {
		log.Println("Error changing username", flowjson.Err)
		return
	}
}
func (p *PgClient) ChangePrivacyDirect(ctx context.Context, flowjson *FlowJSON) {
	_, flowjson.Err = p.Exec(ctx, "UPDATE users SET allow_direct_messages = $1 WHERE user_id = $2", flowjson.Bool, p.UserID)
	if flowjson.Err != nil {
		log.Println("Error changing username", flowjson.Err)
		return
	}
}
func (p *PgClient) ChangePrivacyGroup(ctx context.Context, flowjson *FlowJSON) {
	_, flowjson.Err = p.Exec(ctx, "UPDATE users SET allow_group_invites = $1 WHERE user_id = $2", flowjson.Bool, p.UserID)
	if flowjson.Err != nil {
		log.Println("Error changing username", flowjson.Err)
		return
	}
}

// Blocking user and delete user from duo room
func (p *PgClient) BlockUser(ctx context.Context, flowjson *FlowJSON) {

	p.IsDuoRoomExist(ctx, flowjson)
	if flowjson.Err != nil {
		log.Println(flowjson.Err, "isroomexist err")
		return
	}
	if flowjson.Room != 0 {
		p.DeleteUsersFromRoom(ctx, flowjson)
		if flowjson.Err != nil {
			log.Println(flowjson.Err, "DeleteUsersFromRoom err")
			return
		}
	}

	_, flowjson.Err = flowjson.Tx.Exec(ctx, `INSERT INTO blocked_users (blocked_by_user_id, blocked_user_id)
		VALUES ($1, $2)`, flowjson.Users[0], flowjson.Users[1])
	if flowjson.Err != nil {
		log.Println("Error blocking user", flowjson.Err)
		return
	}
}

// Unblocking user
func (c *PgClient) UnblockUser(ctx context.Context, flowjson *FlowJSON) {
	_, flowjson.Err = flowjson.Tx.Exec(ctx, `DELETE FROM blocked_users 
			WHERE blocked_by_user_id = $1 AND blocked_user_id = $2`, c.UserID, flowjson.Users[1])
	if flowjson.Err != nil {
		log.Println("Error unblocking user", flowjson.Err)
		return
	}
}

func (c *PgClient) Channel() <-chan *FlowJSON {
	return c.Chan
}
func (c *PgClient) ClearChannel() {
	for len(c.Chan) > 0 {
		<-c.Chan
	}
}
