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
		"GetAllRooms":         {GetAllRooms, 1, 1, time.Second},              // initial start
		"GetBlockedUsers":     {GetBlockedUsers, 1, 1, time.Second},          // panel
		"GetMessagesFromRoom": {GetMessagesFromRoom, 1, 1, time.Second},      // from group room and user room
		"FindUsers":           {FindUsers, 1, 1, time.Second},                // panel
		"SendMessage":         {SendMessage, 1, 1, time.Second},              // from group room and user room
		"ChangeUsername":      {ChangeUsername, 1, 1, time.Hour * 24 * 7},    //panel
		"ChangePrivacyDirect": {ChangePrivacyDirect, 1, 1, time.Minute},      //panel
		"ChangePrivacyGroup":  {ChangePrivacyGroup, 1, 1, time.Minute},       //panel
		"CreateDuoRoom":       {CreateDuoRoom, 5, 25, time.Minute * 5},       // from find users panel
		"CreateGroupRoom":     {CreateGroupRoom, 5, 25, time.Minute * 5},     // panel
		"AddUsersToRoom":      {AddUsersToRoom, 5, 25, time.Minute * 5},      // from group room
		"DeleteUsersFromRoom": {DeleteUsersFromRoom, 5, 25, time.Minute * 5}, // from group room if admin
		"BlockUser":           {BlockUser, 5, 25, time.Minute * 5},           // from user room
		"UnblockUser":         {UnblockUser, 5, 25, time.Minute * 5},         // from find users panel
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
func (p *PgClient) GetUserId(ctx context.Context, sub string) (userid uint32, username string, err error) {
	row := p.QueryRow(ctx, "SELECT user_id,user_name FROM users WHERE user_subject=$1", sub)
	err = row.Scan(&userid, &username)
	if err == pgx.ErrNoRows {
		err = p.QueryRow(ctx, "INSERT INTO users (user_subject) VALUES ($1) RETURNING user_id, user_name", sub).Scan(&userid, &username)
		if err != nil {
			return userid, username, fmt.Errorf("failed to create new user: %v", err)
		}
	} else {
		return userid, username, fmt.Errorf("failed to query user: %v", err)
	}
	return userid, username, err
}
