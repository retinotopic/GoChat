package db

import (
	"context"
	"errors"
	"log"
	"strings"
	"unicode"

	json "github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/server/models"
)

func ChangeUsername(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	u, err := models.UnmarshalEvent[models.User](event.Data)
	if err != nil {
		return err
	}
	if len(u.Username) == 0 {
		return errors.New("malformed json")
	}
	username, err := NormalizeString(u.Username)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, "UPDATE users SET user_name = $1 WHERE user_id = $2", username, event.UserId)
	if tag.RowsAffected() == 0 {
		return errors.New("database error: username hasn't changed")
	}
	if err != nil {
		return err
	}
	return err
}
func ChangePrivacyDirect(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	u, err := models.UnmarshalEvent[models.User](event.Data)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, "UPDATE users SET allow_direct_messages = $1 WHERE user_id = $2", u.RoomToggle, event.UserId)
	if tag.RowsAffected() == 0 {
		return errors.New("internal database error, 'allow direct messages' hasn't changed")
	}
	if err != nil {
		return err
	}
	event.Type = 4
	return err
}
func ChangePrivacyGroup(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	u, err := models.UnmarshalEvent[models.User](event.Data)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, "UPDATE users SET allow_group_invites = $1 WHERE user_id = $2", u.RoomToggle, event.UserId)
	if tag.RowsAffected() == 0 {
		return errors.New(" 'allow group invites' hasn't changed")
	}
	if err != nil {
		return err
	}
	event.Type = 4
	return err
}
func FindUsers(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	u, err := models.UnmarshalEvent[models.User](event.Data)
	log.Println(u, "uesers")
	if err != nil {
		return err
	}
	if len(u.Username) == 0 {
		return errors.New("malformed json")
	}
	rows, err := tx.Query(ctx,
		`SELECT user_id,user_name FROM users WHERE user_name ILIKE $1 ORDER BY user_id LIMIT 100`, u.Username+"%")
	if err != nil {
		return err
	}
	resp, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.User])
	if err != nil {
		return err
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	log.Println(resp, "actually uesers")
	return err
}
func GetBlockedUsers(ctx context.Context, tx pgx.Tx, event *models.EventMetadata) error {
	rows, err := tx.Query(ctx,
		`SELECT u.user_name,u.user_id FROM blocked_users b JOIN users u ON u.user_id = b.blocked_user_id
		WHERE blocked_by_user_id = $1 ORDER BY u.user_id`, event.UserId)
	if err != nil {
		return err
	}
	resp, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.User])
	if err != nil {
		return err
	}
	event.Data, err = json.Marshal(resp)
	if err != nil {
		return err
	}
	return err
}
func NormalizeString(input string) (string, error) {
	var builder strings.Builder
	for _, r := range input {
		if unicode.IsLetter(r) && ((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			builder.WriteRune(unicode.ToLower(r))
		}
		if builder.Len() == 30 {
			break
		}
	}
	if builder.Len() == 0 {
		return "", errors.New("malformed name")
	}
	return builder.String(), nil
}
