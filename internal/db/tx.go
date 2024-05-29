package db

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

var once sync.Once

func (c *PostgresClient) TxManage(flowjson *FlowJSON) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	flowjson.Status = "OK"
	fn, ok := c.funcmap[flowjson.Mode]
	if !ok { // if there is no such mode just ignore it
		return
	}

	if fn.ClientOnly {
		if flowjson.Mode == "GetAllRooms" {
			once.Do(func() { fn.Fn(ctx, flowjson) })
		} else {
			fn.Fn(ctx, flowjson)
		}
		if flowjson.Err != nil {
			flowjson.Status = flowjson.Err.Error()
		}
		c.Chan <- *flowjson
		return
	}
	c.txBegin(ctx, flowjson)
	fn.Fn(ctx, flowjson)
	c.txCommit(ctx, flowjson)

	if flowjson.Err != nil {
		flowjson.Status = flowjson.Err.Error()
	}
	c.Chan <- *flowjson
}
func (c *PostgresClient) txBegin(ctx context.Context, flowjson *FlowJSON) {
	flowjson.Tx, flowjson.Err = c.Conn.BeginTx(ctx, pgx.TxOptions{})
}
func (c *PostgresClient) txCommit(ctx context.Context, flowjson *FlowJSON) {
	if flowjson.Err == nil {
		flowjson.Err = flowjson.Tx.Commit(ctx)
		if flowjson.Err != nil {
			flowjson.Status = "internal database error"
			flowjson.Err = flowjson.Tx.Rollback(ctx)
			if flowjson.Err != nil {
				log.Println("ATTENTION Error rolling back transaction:", flowjson.Err)
			}
		}
	} else {
		flowjson.Err = flowjson.Tx.Rollback(ctx)
		if flowjson.Err != nil {
			log.Println("ATTENTION Error rolling back transaction:", flowjson.Err)
			flowjson.Status = "internal database error"
		}
	}
}
