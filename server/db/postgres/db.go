package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/retinotopic/GoChat/server/models"
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

func NewPgClient(ctx context.Context, addr string, lm Limiter) (*PgClient, error) {
	var err error
	pl, err := pgxpool.New(ctx, addr)
	if err != nil {
		return nil, err
	}
	pg := &PgClient{}
	pg.Pool = pl
	pg.Lm = lm
	pg.UserApi = map[string]FuncLimiter{
		"GetAllRooms":         {GetAllRooms, 1, 1, time.Second},
		"ChangeRoomname":      {ChangeRoomname, 1, 1, time.Second},
		"GetBlockedUsers":     {GetBlockedUsers, 1, 1, time.Second},
		"GetMessagesFromRoom": {GetMessagesFromRoom, 1, 1, time.Second},
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
	_, err := p.Exec(ctx, "INSERT INTO users (user_subject,user_name,allow_group_invites,allow_direct_messages) VALUES ($1,$2,true,true)", sub, name)
	return err
}
func (p *PgClient) FuncApi(ctx context.Context, event *models.Event) error {
	fn, ok := p.UserApi[event.Event]
	if ok {
		err := p.Lm.Allow(ctx, event.Event, fn.rate, fn.burst, fn.period)
		if err != nil {
			return err
		}
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
func (p *PgClient) GetUser(ctx context.Context, sub string) (b []byte, userid uint64, err error) {
	var username string
	row := p.QueryRow(ctx, "SELECT user_id,user_name FROM users WHERE user_subject=$1", sub)
	err = row.Scan(&userid, &username)
	if err == pgx.ErrNoRows {
		err = p.QueryRow(ctx, "INSERT INTO users (user_subject) VALUES ($1) RETURNING user_id, user_name", sub).Scan(&userid, &username)
		if err != nil {
			return nil, userid, fmt.Errorf("failed to create new user: %v", err)
		}
	} else {
		return nil, userid, fmt.Errorf("failed to query user: %v", err)
	}
	u := User{UserId: userid, Username: username}
	b, err = json.Marshal(u)
	if err != nil {
		return nil, userid, err
	}
	return b, userid, err
}
