package db

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/retinotopic/GoChat/pkg/str"
)

type FlowJSON struct {
	Mode       string   `json:"Mode"`
	Message    string   `json:"Message" db:"payload"`
	Users      []uint32 `json:"Users"`
	User       uint32   `json:"User" db:"user_id"`
	Room       uint32   `json:"Room" db:"room_id"`
	Name       string   `json:"Name" db:"username"`
	Message_id string   `json:"Offset"`
	ErrorMsg   string   `json:"ErrorMsg"`
	Bool       bool     `json:"Bool"`
}

type PgClient struct {
	Sub              string
	Name             string
	UserID           uint32
	Mutex            sync.Mutex
	RoomsPagination  []uint32
	RoomsCount       uint8 // no more than 250
	PaginationOffset uint8
	funcmap          map[string]funcapi
	Chan             chan FlowJSON
	Once             sync.Once
	*pgxpool.Pool
}
type funcapi = func(context.Context, *FlowJSON) error

func (p *PgClient) FuncApi(ctx context.Context, cancelFunction context.CancelFunc, fj *FlowJSON) {
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
func (p *PgClient) SendMessage(ctx context.Context, flowjson *FlowJSON) error {
	_, err := p.Exec(ctx, `INSERT INTO messages (payload,user_id,room_id) VALUES ($1,$2,$3)`, flowjson.Message, p.UserID, flowjson.Room)
	if err != nil {
		return err
	}
	return err
}
func (p *PgClient) ChangeUsername(ctx context.Context, flowjson *FlowJSON) error {
	username := str.NormalizeString(flowjson.Name)
	_, err := p.Exec(ctx, "UPDATE users SET username = $1 WHERE user_id = $2", username, p.UserID)
	if err != nil {
		return err
	}
	return err
}
func (p *PgClient) ChangePrivacyDirect(ctx context.Context, flowjson *FlowJSON) error {
	_, err := p.Exec(ctx, "UPDATE users SET allow_direct_messages = $1 WHERE user_id = $2", flowjson.Bool, p.UserID)
	if err != nil {
		return err
	}
	return err
}
func (p *PgClient) ChangePrivacyGroup(ctx context.Context, flowjson *FlowJSON) error {
	_, err := p.Exec(ctx, "UPDATE users SET allow_group_invites = $1 WHERE user_id = $2", flowjson.Bool, p.UserID)
	if err != nil {
		return err
	}
	return err
}

func (c *PgClient) Channel() <-chan FlowJSON {
	return c.Chan
}
func (c *PgClient) ClearChannel() {
	for len(c.Chan) > 0 {
		<-c.Chan
	}
}
