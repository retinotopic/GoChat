package db

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

var once sync.Once

func (p *PgClient) TxManage(flowjson *FlowJSON) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	flowjson.Status = "OK"
	fn, ok := p.funcmap[flowjson.Mode]
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
			p.Chan <- *flowjson
		}
		return
	}
	p.txBegin(ctx, flowjson)
	fn.Fn(ctx, flowjson)
	p.txCommit(ctx, flowjson)

	if flowjson.Err != nil {
		flowjson.Status = flowjson.Err.Error()
		log.Println(flowjson.Err)
	}
	p.Chan <- *flowjson
}
func (p *PgClient) txBegin(ctx context.Context, flowjson *FlowJSON) {
	flowjson.Tx, flowjson.Err = p.BeginTx(ctx, pgx.TxOptions{})
}
func (p *PgClient) txCommit(ctx context.Context, flowjson *FlowJSON) {
	if flowjson.Err == nil {
		flowjson.Err = flowjson.Tx.Commit(ctx)
		if flowjson.Err != nil {
			flowjson.Status = "internal database error"
			err := flowjson.Tx.Rollback(ctx)
			if err != nil {
				log.Println("ATTENTION Error rolling back transaction:", flowjson.Err)
			}
		}
	} else {
		err := flowjson.Tx.Rollback(ctx)
		if err != nil {
			log.Println("ATTENTION Error rolling back transaction:", flowjson.Err)
			flowjson.Status = "internal database error"
		}
	}
}
