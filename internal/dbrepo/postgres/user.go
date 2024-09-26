package db

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/internal/models"
	"github.com/retinotopic/GoChat/pkg/str"
)

type User struct {
	UserId   uint32 `json:"UserId"`
	Username string `json:"Username" `
	Bool     bool   `json:"Bool" `
	tx       pgx.Tx `json:"-"`
}

func ChangeUsername(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	u := &User{}
	err := json.Unmarshal(event.Data, u)
	if err != nil {
		return err
	}
	if len(u.Username) == 0 {
		return fmt.Errorf("malformed json")
	}
	name := str.NormalizeString(u.Username)
	_, err = u.tx.Exec(ctx, "UPDATE users SET username = $1 WHERE user_id = $2", name, event.UserId)
	if err != nil {
		return err
	}
	return err
}
func ChangePrivacyDirect(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	u := &User{}
	err := json.Unmarshal(event.Data, u)
	if err != nil {
		return err
	}
	_, err = u.tx.Exec(ctx, "UPDATE users SET allow_direct_messages = $1 WHERE user_id = $2", u.Bool, event.UserId)
	if err != nil {
		return err
	}
	return err
}
func ChangePrivacyGroup(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	u := &User{}
	err := json.Unmarshal(event.Data, u)
	if err != nil {
		return err
	}
	_, err = u.tx.Exec(ctx, "UPDATE users SET allow_group_invites = $1 WHERE user_id = $2", u.Bool, event.UserId)
	if err != nil {
		return err
	}
	return err
}
func FindUsers(ctx context.Context, tx pgx.Tx, event *models.Event) error {
	u := &User{}
	err := json.Unmarshal(event.Data, u)
	if err != nil {
		return err
	}
	if len(u.Username) == 0 {
		return fmt.Errorf("malformed json")
	}
	rows, err := tx.Query(ctx,
		`SELECT user_id,username FROM users WHERE username ILIKE $1 LIMIT 20`, u.Username+"%")
	if err != nil {
		return err
	}
	resp, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[User])
	if err != nil {
		return err
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	return err
}
