package db

import (
	"context"
	"log"
)

func (c *PostgresClient) TxManage(flowjson *FlowJSON) {
	flowjson.Status = "OK"
	fn, ok := c.funcmap[flowjson.Mode]
	if !ok { // if there is no such mode just ignore it
		return
	}
	if flowjson.SenderOnly {
		fn(flowjson)
		if flowjson.Err != nil {
			flowjson.Status = flowjson.Err.Error()
		}
		c.Chan <- *flowjson
		return
	}
	c.txBegin(flowjson)
	fn(flowjson)
	c.txCommit(flowjson)

	if flowjson.Err != nil {
		flowjson.Status = flowjson.Err.Error()
	}
	c.Chan <- *flowjson
}
func (c *PostgresClient) txBegin(flowjson *FlowJSON) {
	flowjson.Tx, flowjson.Err = flowjson.Tx.Begin(context.Background())
}
func (c *PostgresClient) txCommit(flowjson *FlowJSON) {
	if flowjson.Err == nil {
		flowjson.Err = flowjson.Tx.Commit(context.Background())
		if flowjson.Err != nil {
			flowjson.Status = "internal database error"
			flowjson.Err = flowjson.Tx.Rollback(context.Background())
			if flowjson.Err != nil {
				log.Println("ATTENTION Error rolling back transaction:", flowjson.Err)
			}
		}
	} else {
		flowjson.Err = flowjson.Tx.Rollback(context.Background())
		if flowjson.Err != nil {
			log.Println("ATTENTION Error rolling back transaction:", flowjson.Err)
			flowjson.Status = "internal database error"
		}
	}
}
