package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/retinotopic/GoChat/internal/models"
)

func (p *PgClient) TxManage(next func(ctx context.Context, tx pgx.Tx, fj *models.Flowjson) error) funcapi {
	return func(ctx context.Context, fj *models.Flowjson) error {
		tx, err := p.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}
		defer func() {
			tx.Rollback(ctx)
		}()
		err = next(ctx, tx, fj)
		if err != nil {
			return err
		}
		err = tx.Commit(ctx)
		if err != nil {
			return err
		}
		return nil
	}
}
