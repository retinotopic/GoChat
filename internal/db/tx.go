package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func (p *PgClient) TxManage(next funcapi) funcapi {
	return func(ctx context.Context, fj *FlowJSON) {
		p.txBegin(ctx, fj)
		next(ctx, fj)
		p.txCommit(ctx, fj)
	}
}
func (p *PgClient) txBegin(ctx context.Context, flowjson *FlowJSON) {
	flowjson.Tx, flowjson.Err = p.BeginTx(ctx, pgx.TxOptions{})
}
func (p *PgClient) txCommit(ctx context.Context, flowjson *FlowJSON) {
	if flowjson.Err == nil {
		flowjson.Err = flowjson.Tx.Commit(ctx)
		if flowjson.Err != nil {
			flowjson.ErrorMsg = "internal database error"
			err := flowjson.Tx.Rollback(ctx)
			if err != nil {
				log.Println("ATTENTION Error rolling back transaction:", flowjson.Err)
			}
		}
	} else {
		err := flowjson.Tx.Rollback(ctx)
		if err != nil {
			log.Println("ATTENTION Error rolling back transaction:", flowjson.Err)
			flowjson.ErrorMsg = "internal database error"
		}
	}
}
