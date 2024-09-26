package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/internal/models"
	"github.com/retinotopic/GoChat/pkg/str"
)

type User struct {
	UserId uint32 `json:"UserId"`
	Name   string `json:"Name" `
	Bool   bool   `json:"Bool" `
	tx     pgx.Tx `json:"-"`
}

func (u *User) ChangeUsername(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	name := str.NormalizeString(u.Name)
	_, err := u.tx.Exec(ctx, "UPDATE users SET username = $1 WHERE user_id = $2", name, u.UserId)
	if err != nil {
		return err
	}
	return err
}
func (u *User) ChangePrivacyDirect(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	_, err := u.tx.Exec(ctx, "UPDATE users SET allow_direct_messages = $1 WHERE user_id = $2", u.Bool, u.UserId)
	if err != nil {
		return err
	}
	return err
}
func (u *User) ChangePrivacyGroup(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	_, err := u.tx.Exec(ctx, "UPDATE users SET allow_group_invites = $1 WHERE user_id = $2", u.Bool, u.UserId)
	if err != nil {
		return err
	}
	return err
}
func (p *PgClient) FindUsers(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	rows, err := p.Query(ctx,
		`SELECT user_id,username FROM users WHERE username ILIKE $1 LIMIT 20`, event.Payload+"%")
	if err != nil {
		return err
	}
	fjarr, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[User])

	if err != nil {
		return err
	}
	return err
}
