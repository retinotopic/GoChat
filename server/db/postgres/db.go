package db

import (
	"context"
	"fmt"
	"log"
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
	fn     func(context.Context, pgx.Tx, *models.EventMetadata) error
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
		"Get All Rooms":          {GetAllRooms, 1, 1, time.Second},
		"Change Room Name":       {ChangeRoomname, 1, 1, time.Second},
		"Get Blocked Users":      {GetBlockedUsers, 1, 1, time.Second},
		"Get Messages From Room": {GetMessagesFromRoom, 1, 1, time.Second},
		"Find Users":             {FindUsers, 1, 1, time.Second},
		"Send Message":           {SendMessage, 1, 1, time.Second},
		"Change Username":        {ChangeUsername, 1, 1, time.Hour * 24 * 7},
		"Change Privacy Direct":  {ChangePrivacyDirect, 1, 1, time.Second},
		"Change Privacy Group":   {ChangePrivacyGroup, 1, 1, time.Second},
		"Create Duo Room":        {CreateDuoRoom, 5, 25, time.Minute * 5},
		"Create Group Room":      {CreateGroupRoom, 5, 25, time.Minute * 5},
		"Add Users To Room":      {AddUsersToRoom, 5, 25, time.Minute * 5},
		"Delete Users From Room": {DeleteUsersFromRoom, 5, 25, time.Minute * 5},
		"Block User":             {BlockUser, 5, 25, time.Minute * 5},
		"Unblock User":           {UnblockUser, 5, 25, time.Minute * 5},
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
		err := p.Lm.Allow(ctx, event.Event, fn.rate, fn.burst, fn.period)
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
	} else {
		log.Println("i hate this")

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
