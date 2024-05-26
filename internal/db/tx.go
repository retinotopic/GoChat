package db

import (
	"context"
	"log"
	"time"
)

func (c *PostgresClient) TxManage(flowjson *FlowJSON) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	flowjson.Status = "OK"
	fn, ok := c.funcmap[flowjson.Mode]
	if !ok { // if there is no such mode just ignore it
		return
	}
	if flowjson.SenderOnly {
		fn(ctx, flowjson)
		if flowjson.Err != nil {
			flowjson.Status = flowjson.Err.Error()
		}
		c.Chan <- *flowjson
		return
	}
	c.txBegin(ctx, flowjson)
	fn(ctx, flowjson)
	c.txCommit(ctx, flowjson)

	if flowjson.Err != nil {
		flowjson.Status = flowjson.Err.Error()
	}
	c.Chan <- *flowjson
}
func (c *PostgresClient) txBegin(ctx context.Context, flowjson *FlowJSON) {
	flowjson.Tx, flowjson.Err = flowjson.Tx.Begin(ctx)
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
