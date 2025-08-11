package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/retinotopic/GoChat/server/models"
	"github.com/valkey-io/valkey-go/valkeylimiter"
)

type PgClient struct {
	*pgxpool.Pool
	Limiter valkeylimiter.RateLimiterClient
	IsDebug bool
	UserApi map[string]FuncLimiter
}

type FuncLimiter struct {
	fn     func(context.Context, pgx.Tx, *models.EventMetadata) error
	limit  int
	window time.Duration
}

func (p *PgClient) Allow(ctx context.Context, key string, limit int, window time.Duration) (err error) {
	if p.IsDebug {
		return nil
	}
	options := valkeylimiter.WithCustomRateLimit(limit, window)
	res, err := p.Limiter.Allow(ctx, key, options)
	if err != nil {
		return err
	}
	if res.Allowed {
		return nil
	}
	return errors.New("limit exceeded, retry after " + strconv.FormatUint(uint64(res.Remaining), 10))
}

func NewPgClient(ctx context.Context, addr string, lm valkeylimiter.RateLimiterClient, isdebug bool) (*PgClient, error) {
	var err error
	pl, err := pgxpool.New(ctx, addr)
	if err != nil {
		return nil, err
	}
	pg := &PgClient{}
	pg.Pool = pl
	pg.IsDebug = isdebug
	pg.Limiter = lm
	pg.UserApi = map[string]FuncLimiter{
		"Get All Rooms":          {GetUserRooms, 1, time.Second},
		"Change Room Name":       {ChangeRoomname, 1, time.Second},
		"Get Blocked Users":      {GetBlockedUsers, 1, time.Second},
		"Get Messages From Room": {GetMessagesFromRoom, 1, time.Second},
		"Find Users":             {FindUsers, 1, time.Second},
		"Send Message":           {SendMessage, 1, time.Second},
		"Change Username":        {ChangeUsername, 1, time.Hour * 24 * 7},
		"Change Privacy Direct":  {ChangePrivacyDirect, 1, time.Second},
		"Change Privacy Group":   {ChangePrivacyGroup, 1, time.Second},
		"Create Duo Room":        {CreateDuoRoom, 20, time.Minute * 5},
		"Create Group Room":      {CreateGroupRoom, 20, time.Minute * 5},
		"Add Users To Room":      {AddUsersToRoom, 20, time.Minute * 5},
		"Delete Users From Room": {DeleteUsersFromRoom, 20, time.Minute * 5},
		"Block User":             {BlockUser, 20, time.Minute * 5},
		"Unblock User":           {UnblockUser, 20, time.Minute * 5},
	}
	return pg, nil
}

//	func (p *PgClient) NewUser(ctx context.Context, sub, name string) error {
//		if strings.ContainsAny(name, " \t\n") {
//			return errors.New("contains spaces")
//		}
//		_, err := p.Exec(ctx, "INSERT INTO users (user_subject,user_name,allow_group_invites,allow_direct_messages) VALUES ($1,$2,true,true)", sub, name)
//		return err
//	}

func (p *PgClient) FuncApi(ctx context.Context, event *models.EventMetadata) error {
	fn, ok := p.UserApi[event.Event]
	if ok {
		err := p.Allow(ctx, event.Event, fn.limit, fn.window)
		if err != nil {
			return err
		}
		tx, err := p.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
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
	}
	u := models.User{UserId: userid, Username: username}
	b, err = json.Marshal(u)
	if err != nil {
		return nil, userid, err
	}
	return b, userid, err
}
