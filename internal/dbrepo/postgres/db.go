package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/retinotopic/GoChat/internal/models"
)

type PgClient struct {
	*pgxpool.Pool
	Lm      Limiter
	UserApi map[string]FuncLimiter
}
type Limiter interface {
	Allow(ctx context.Context, key string, rate int, burst int, period time.Duration) (err error)
}
type FuncLimiter struct {
	fn     func(context.Context, pgx.Tx, *models.Event) error
	rate   int
	burst  int
	period time.Duration
}

func NewPgClient(ctx context.Context, addr string) (*PgClient, error) {
	var err error
	pl, err := pgxpool.New(ctx, addr)
	if err != nil {
		return nil, err
	}
	pg := &PgClient{}
	pg.Pool = pl
	pg.UserApi = map[string]FuncLimiter{
		"GetAllRoomsIds":      {GetAllRoomsIds, 1, 1, time.Second},
		"GetMessagesFromRoom": {GetMessagesFromRoom, 1, 1, time.Second},
		"GetNextRooms":        {GetNextRooms, 1, 9, time.Second},
		"FindUsers":           {FindUsers, 1, 1, time.Second},
		"SendMessage":         {SendMessage, 1, 1, time.Second},
		"ChangeUsername":      {ChangeUsername, 1, 1, time.Hour * 24 * 7},
		"ChangePrivacyDirect": {ChangePrivacyDirect, 1, 1, time.Minute},
		"ChangePrivacyGroup":  {ChangePrivacyGroup, 1, 1, time.Minute},
		"CreateDuoRoom":       {CreateDuoRoom, 5, 25, time.Minute * 5},
		"CreateGroupRoom":     {CreateGroupRoom, 5, 25, time.Minute * 5},
		"AddUsersToRoom":      {AddUsersToRoom, 5, 25, time.Minute * 5},
		"DeleteUsersFromRoom": {DeleteUsersFromRoom, 5, 25, time.Minute * 5},
		"BlockUser":           {BlockUser, 5, 25, time.Minute * 5},
		"UnblockUser":         {UnblockUser, 5, 25, time.Minute * 5},
	}
	return pg, nil
}
func (p *PgClient) NewUser(ctx context.Context, sub, name string) error {
	if strings.ContainsAny(name, " \t\n") {
		return errors.New("contains spaces")
	}
	_, err := p.Exec(ctx, "INSERT INTO users (subject,name,allow_group_invites,allow_direct_messages) VALUES ($1,$2,true,true)", sub, name)
	return err
}
func (p *PgClient) FuncApi(ctx context.Context, event *models.Event) error {
	fn, ok := p.UserApi[event.Event]
	if ok {
		tx, err := p.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}
		defer func() {
			tx.Rollback(ctx)
		}()
		err = fn.fn(ctx, tx, event)
		if err != nil {
			return err
		}
		err = tx.Commit(ctx)
		if err != nil {
			return err
		}

	}

	return nil

}
func (p *PgClient) GetUserId(ctx context.Context, sub string) (userid uint32, err error) {
	row := p.QueryRow(ctx, "SELECT user_id,name FROM users WHERE subject=$1", sub)
	err = row.Scan(&userid)
	if err == pgx.ErrNoRows {
		err = p.QueryRow(ctx, "INSERT INTO users (subject) VALUES ($1) RETURNING user_id, username", sub).Scan(&userid)
		if err != nil {
			return userid, fmt.Errorf("failed to create new user: %v", err)
		}
	} else {
		return userid, fmt.Errorf("failed to query user: %v", err)
	}
	return userid, err
}
