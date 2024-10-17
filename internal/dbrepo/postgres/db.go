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
}
type funcapi = func(context.Context, pgx.Tx, *models.Event) error

func NewPgClient(ctx context.Context, addr string) (*PgClient, error) {
	var err error
	pl, err := pgxpool.New(ctx, addr)
	if err != nil {
		return nil, err
	}
	pg := &PgClient{}
	pg.Pool = pl
	pg.UserApi = map[string]funcapi{
		"GetAllRoomsIds":      GetAllRoomsIds,
		"GetMessagesFromRoom": GetMessagesFromRoom,
		"GetNextRooms":        GetNextRooms,
		"FindUsers":           FindUsers,
		"SendMessage":         SendMessage,         // 1 second, burst 10
		"Changename":          ChangeUsername,      // 168 hours, burst 10
		"ChangePrivacyDirect": ChangePrivacyDirect, // 1 minute, burst 10
		"ChangePrivacyGroup":  ChangePrivacyGroup,  // 1 minute, burst 10
		"CreateDuoRoom":       CreateDuoRoom,       // 1 minute, burst 10
		"CreateGroupRoom":     CreateGroupRoom,     // 1 minute, burst 10
		"AddUsersToRoom":      AddUsersToRoom,      // 168 hours, burst 10
		"DeleteUsersFromRoom": DeleteUsersFromRoom, // 168 hours, burst 10
		"BlockUser":           BlockUser,           // 168 hours, burst 10
		"UnblockUser":         UnblockUser,         // 168 hours, burst 10
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
func (p *PgClient) GetUser(ctx context.Context, sub string) (userid uint32, err error) {
	row := p.QueryRow(ctx, "SELECT user_id,name FROM users WHERE subject=$1", sub)
	err = row.Scan(&userid)
	if err == pgx.ErrNoRows {
		err = p.QueryRow(ctx, "INSERT INTO users (subject, name) VALUES ($1, $2) RETURNING user_id, name", sub, fmt.Sprintf("user%v", new(maphash.Hash).Sum64())).Scan(&userid)
		if err != nil {
			return userid, fmt.Errorf("failed to create new user: %v", err)
		}
	} else {
		return userid, fmt.Errorf("failed to query user: %v", err)
	}
	return userid, err
}
