package db

import (
	"context"
	"errors"
	"fmt"
	"hash/maphash"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/retinotopic/GoChat/internal/models"
)

type PgClient struct {
	*pgxpool.Pool
	UserApi map[string]funcapi
	actions [][]string
}
type funcapi = func(context.Context, pgx.Tx, *models.Event) error

func (p *PgClient) NewPgClient(ctx context.Context, addr string) (*PgClient, error) {
	var err error
	pl, err := pgxpool.New(ctx, addr)
	if err != nil {
		return nil, err
	}
	pg := &PgClient{}
	pg.Pool = pl
	pg.actions = [][]string{{"CreateGroupRoom", "CreateDuoRoom", "AddUserToRoom"}, {"DeleteUsersFromRoom", "BlockUser"}, {"SendMessage"}}

	pg.UserApi = map[string]funcapi{
		"GetAllRooms":         GetAllRooms,
		"GetMessagesFromRoom": GetMessagesFromRoom,
		"GetNextRooms":        GetNextRooms,
		"FindUsers":           FindUsers,
		"SendMessage":         SendMessage,
		"Changename":          ChangeUsername,
		"ChangePrivacyDirect": ChangePrivacyDirect,
		"ChangePrivacyGroup":  ChangePrivacyGroup,
		"CreateDuoRoom":       CreateDuoRoom,
		"CreateGroupRoom":     CreateRoom,
		"AddUsersToRoom":      AddUsersToRoom,
		"DeleteUsersFromRoom": DeleteUsersFromRoom,
		"BlockUser":           BlockUser,
		"UnblockUser":         CreateDuoRoom,
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
func (p *PgClient) FuncApi(ctx context.Context, cancelFunc context.CancelFunc, event *models.Event) error {
	defer cancelFunc()
	fn, ok := p.UserApi[event.Mode]
	if ok {
		tx, err := p.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}
		defer func() {
			tx.Rollback(ctx)
		}()
		err = fn(ctx, tx, event)
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
func (p *PgClient) GetUser(ctx context.Context, sub string) (userid uint32, name string, err error) {
	row := p.QueryRow(ctx, "SELECT user_id,name FROM users WHERE subject=$1", sub)
	err = row.Scan(&userid, &name)
	if err == pgx.ErrNoRows {
		err = p.QueryRow(ctx, "INSERT INTO users (subject, name) VALUES ($1, $2) RETURNING user_id, name", sub, fmt.Sprintf("user%v", new(maphash.Hash).Sum64())).Scan(&userid, &name)
		if err != nil {
			return userid, name, fmt.Errorf("failed to create new user: %v", err)
		}
	} else {
		return userid, name, fmt.Errorf("failed to query user: %v", err)
	}
	return userid, name, err
}

func (p *PgClient) PubSubActions(id int) []string {
	if id >= len(p.actions) {
		return []string{}
	}
	return p.actions[id]
}
