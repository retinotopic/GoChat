package db

import (
	"context"
	"errors"
	"fmt"
	"hash/maphash"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/retinotopic/GoChat/internal/models"
)

type Pool struct {
	*pgxpool.Pool
}

func NewPool(ctx context.Context, addr string) (*Pool, error) {
	var err error
	pl, err := pgxpool.New(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &Pool{Pool: pl}, nil
}
func (p *Pool) GetClient(ctx context.Context, sub string) (*PgClient, error) {
	row := p.QueryRow(ctx, "SELECT user_id,username FROM users WHERE subject=$1", sub)
	var name string
	var userid uint32
	err := row.Scan(&userid, &name)

	if err == pgx.ErrNoRows {
		err = p.QueryRow(ctx, "INSERT INTO users (subject, username) VALUES ($1, $2) RETURNING user_id, username", sub, fmt.Sprintf("user%v", new(maphash.Hash).Sum64())).Scan(&userid, &name)
		if err != nil {
			return nil, fmt.Errorf("failed to create new user: %v", err)
		}
	} else {
		return nil, fmt.Errorf("failed to query user: %v", err)
	}
	pg := &PgClient{}
	pg.Sub = sub
	pg.UserID = userid
	pg.Name = name
	pg.Chan = make(chan models.Flowjson, 1000)
	pg.actions = [][]string{{"CreateGroupRoom", "CreateDuoRoom", "AddUserToRoom"}, {"DeleteUsersFromRoom", "BlockUser"}, {"SendMessage"}}

	pg.funcmap = map[string]funcapi{
		"GetAllRooms":         pg.GetAllRooms,
		"GetMessagesFromRoom": pg.GetMessagesFromRoom,
		"GetNextRooms":        pg.GetNextRooms,
		"FindUsers":           pg.FindUsers,
		"SendMessage":         pg.SendMessage,
		"ChangeUsername":      pg.ChangeUsername,
		"ChangePrivacyDirect": pg.ChangePrivacyDirect,
		"ChangePrivacyGroup":  pg.ChangePrivacyGroup,
		"CreateDuoRoom":       pg.TxManage(pg.CreateDuoRoom),
		"CreateGroupRoom":     pg.TxManage(pg.CreateDuoRoom),
		"AddUsersToRoom":      pg.TxManage(pg.CreateDuoRoom),
		"DeleteUsersFromRoom": pg.TxManage(pg.CreateDuoRoom),
		"BlockUser":           pg.TxManage(pg.CreateDuoRoom),
		"UnblockUser":         pg.TxManage(pg.CreateDuoRoom),
	}
	return pg, nil
}
func (p *Pool) NewUser(ctx context.Context, sub, username string) error {
	if strings.ContainsAny(username, " \t\n") {
		return errors.New("contains spaces")
	}
	_, err := p.Exec(ctx, "INSERT INTO users (subject,username,allow_group_invites,allow_direct_messages) VALUES ($1,$2,true,true)", sub, username)
	return err
}
