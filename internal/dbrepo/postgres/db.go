package db

import (
	"context"
	"sync"

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
