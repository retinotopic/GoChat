package db

import (
	"context"
	"fmt"
	"hash/maphash"
	"sync"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/retinotopic/GoChat/internal/models"
	"github.com/retinotopic/GoChat/pkg/str"
)

type PgClient struct {
	Sub              string
	Name             string
	UserID           uint32
	NRMutex          sync.Mutex
	GRMutex          sync.Mutex
	RoomsPagination  []uint32
	RoomsCount       uint8 // no more than 250
	PaginationOffset uint8
	funcmap          map[string]funcapi
	Chan             chan models.Flowjson
	ReOnce           bool
	actions          [][]string
	*pgxpool.Pool
}

func (p *PgClient) GetClient(ctx context.Context, sub string) error {
	row := p.QueryRow(ctx, "SELECT user_id,username FROM users WHERE subject=$1", sub)
	var name string
	var userid uint32
	err := row.Scan(&userid, &name)

	if err == pgx.ErrNoRows {
		err = p.QueryRow(ctx, "INSERT INTO users (subject, username) VALUES ($1, $2) RETURNING user_id, username", sub, fmt.Sprintf("user%v", new(maphash.Hash).Sum64())).Scan(&userid, &name)
		if err != nil {
			return fmt.Errorf("failed to create new user: %v", err)
		}
	} else {
		return fmt.Errorf("failed to query user: %v", err)
	}
	p.Sub = sub
	p.UserID = userid
	p.Name = name
	p.Chan = make(chan models.Flowjson, 1000)
	p.actions = [][]string{{"CreateGroupRoom", "CreateDuoRoom", "AddUserToRoom"}, {"DeleteUsersFromRoom", "BlockUser"}, {"SendMessage"}}

	p.funcmap = map[string]funcapi{
		"GetAllRooms":         p.GetAllRooms,
		"GetMessagesFromRoom": p.GetMessagesFromRoom,
		"GetNextRooms":        p.GetNextRooms,
		"FindUsers":           p.FindUsers,
		"SendMessage":         p.SendMessage,
		"ChangeUsername":      p.ChangeUsername,
		"ChangePrivacyDirect": p.ChangePrivacyDirect,
		"ChangePrivacyGroup":  p.ChangePrivacyGroup,
		"CreateDuoRoom":       p.TxManage(p.CreateDuoRoom),
		"CreateGroupRoom":     p.TxManage(p.CreateDuoRoom),
		"AddUsersToRoom":      p.TxManage(p.CreateDuoRoom),
		"DeleteUsersFromRoom": p.TxManage(p.CreateDuoRoom),
		"BlockUser":           p.TxManage(p.CreateDuoRoom),
		"UnblockUser":         p.TxManage(p.CreateDuoRoom),
	}
	return nil
}

type funcapi = func(context.Context, *models.Flowjson) error

func (p *PgClient) FuncApi(ctx context.Context, cancelFunction context.CancelFunc, fj *models.Flowjson) {
	defer cancelFunction()
	fn, ok := p.funcmap[fj.Mode]
	if ok {
		err := fn(ctx, fj)
		if err != nil {
			fj.ErrorMsg = err.Error()
			p.Chan <- *fj
		}
	}
}
func (p *PgClient) SendMessage(ctx context.Context, flowjson *models.Flowjson) error {
	_, err := p.Exec(ctx, `INSERT INTO messages (payload,user_id,room_id) VALUES ($1,$2,$3)`, flowjson.Message, p.UserID, flowjson.Room)
	if err != nil {
		return err
	}
	return err
}
func (p *PgClient) ChangeUsername(ctx context.Context, flowjson *models.Flowjson) error {
	username := str.NormalizeString(flowjson.Name)
	_, err := p.Exec(ctx, "UPDATE users SET username = $1 WHERE user_id = $2", username, p.UserID)
	if err != nil {
		return err
	}
	return err
}
func (p *PgClient) ChangePrivacyDirect(ctx context.Context, flowjson *models.Flowjson) error {
	_, err := p.Exec(ctx, "UPDATE users SET allow_direct_messages = $1 WHERE user_id = $2", flowjson.Bool, p.UserID)
	if err != nil {
		return err
	}
	return err
}
func (p *PgClient) ChangePrivacyGroup(ctx context.Context, flowjson *models.Flowjson) error {
	_, err := p.Exec(ctx, "UPDATE users SET allow_group_invites = $1 WHERE user_id = $2", flowjson.Bool, p.UserID)
	if err != nil {
		return err
	}
	return err
}

func (c *PgClient) Channel() <-chan models.Flowjson {
	return c.Chan
}
func (c *PgClient) ClearChannel() {
	for len(c.Chan) > 0 {
		<-c.Chan
	}
}

func (c *PgClient) PubSubActions(id int) []string {
	if id >= len(c.actions) {
		return []string{}
	}
	return c.actions[id]
}
