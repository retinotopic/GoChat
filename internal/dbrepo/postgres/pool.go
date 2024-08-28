package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/retinotopic/GoChat/internal/models"
)

type Pool struct {
	Pl *pgxpool.Pool
}

func NewPool(ctx context.Context, addr string) (*Pool, error) {
	var err error
	pl, err := pgxpool.New(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &Pool{Pl: pl}, nil
}

func (p *Pool) NewUser(ctx context.Context, sub, username string) error {
	if strings.ContainsAny(username, " \t\n") {
		return errors.New("contains spaces")
	}
	_, err := p.Pl.Exec(ctx, "INSERT INTO users (subject,username,allow_group_invites,allow_direct_messages) VALUES ($1,$2,true,true)", sub, username)
	return err
}
func (p *Pool) GetClient(ctx context.Context, sub string) (*PgClient, error) {
	row := p.Pl.QueryRow(ctx, "SELECT user_id,username FROM users WHERE subject=$1", sub)
	var name string
	var userid uint32
	err := row.Scan(&userid, &name)
	if err == pgx.ErrNoRows {
		err = p.Pl.QueryRow(ctx, "INSERT INTO users (subject, username) VALUES ($1, $2) RETURNING user_id, username", sub, "New User").Scan(&userid, &name)
		if err != nil {
			return nil, fmt.Errorf("failed to create new user: %v", err)
		}
	} else {
		return nil, fmt.Errorf("failed to query user: %v", err)
	}
	pc := &PgClient{
		Sub:     sub,
		UserID:  userid,
		Name:    name,
		Chan:    make(chan models.Flowjson, 1000),
		actions: [][]string{{"CreateGroupRoom", "CreateDuoRoom", "AddUserToRoom"}, {"DeleteUsersFromRoom", "BlockUser"}, {"SendMessage"}},
	}
	pc.funcmap = map[string]funcapi{
		"GetAllRooms":         pc.GetAllRooms,
		"GetMessagesFromRoom": pc.GetMessagesFromRoom,
		"GetNextRooms":        pc.GetNextRooms,
		"FindUsers":           pc.FindUsers,
		"SendMessage":         pc.SendMessage,
		"ChangeUsername":      pc.ChangeUsername,
		"ChangePrivacyDirect": pc.ChangePrivacyDirect,
		"ChangePrivacyGroup":  pc.ChangePrivacyGroup,
		"CreateDuoRoom":       pc.TxManage(pc.CreateDuoRoom),
		"CreateGroupRoom":     pc.TxManage(pc.CreateDuoRoom),
		"AddUsersToRoom":      pc.TxManage(pc.CreateDuoRoom),
		"DeleteUsersFromRoom": pc.TxManage(pc.CreateDuoRoom),
		"BlockUser":           pc.TxManage(pc.CreateDuoRoom),
		"UnblockUser":         pc.TxManage(pc.CreateDuoRoom),
	}
	return pc, nil
}
