package db

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
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

func (p *Pool) NewUser(ctx context.Context, sub, username string) error {
	if strings.ContainsAny(username, " \t\n") {
		return errors.New("contains spaces")
	}
	_, err := p.Exec(ctx, "INSERT INTO users (subject,username,allow_group_invites,allow_direct_messages) VALUES ($1,$2,true,true)", sub, username)
	return err
}
func (p *Pool) NewClient(ctx context.Context, sub string) (*PgClient, error) {
	// check if user exists
	row := p.QueryRow(ctx, "SELECT user_id,username FROM users WHERE subject=$1", sub)
	var name string
	var userid uint32
	err := row.Scan(&userid, &name)
	if err != nil {
		return nil, err
	}
	pc := &PgClient{
		Sub:    sub,
		UserID: userid,
		Name:   name,
		Chan:   make(chan *FlowJSON, 1000),
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