package pubsub

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"

	"github.com/fasthttp/websocket"
	"github.com/retinotopic/GoChat/internal/logger"
	"github.com/retinotopic/GoChat/internal/middleware"
	"github.com/retinotopic/GoChat/internal/models"
	"github.com/retinotopic/GoChat/pkg/wsutils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Connector struct {
	GetDB func(context.Context, string) (Databaser, error)
	GetPS func(context.Context) (PubSuber, error)
	Log   logger.Logger
}

func (c *Connector) Connect(w http.ResponseWriter, r *http.Request) {

	sub := middleware.GetUser(r.Context())
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.Log.Error("upgrade to websocket err", err)
		return
	}

	db, err := c.GetDB(r.Context(), sub)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
	}
	pb, err := c.GetPS(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	p := pubsub{conn: conn, Pb: pb, Db: db, Log: c.Log, writeCh: make(chan models.Flowjson, 500)}
	p.WsHandle()
}

type Databaser interface {
	FuncApi(context.Context, context.CancelFunc, *models.Flowjson)
	PubSubActions(int) []string
	Channel() <-chan models.Flowjson
}
type PubSuber interface {
	Unsubscribe(context.Context, ...string) error
	Subscribe(context.Context, ...string) error
	Publish(context.Context, string, interface{}) error
	Channel() chan models.Flowjson
}

// Publish||Subscribe Service
type pubsub struct {
	conn    *websocket.Conn
	writeCh chan models.Flowjson
	Pb      PubSuber
	Db      Databaser
	Log     logger.Logger
	errch   chan bool
}

// conn, err := upgrader.Upgrade(w, r, nil)

func (p *pubsub) WsHandle() {
	p.errch = make(chan bool, 10)
	defer func() {
		p.conn.Close()
		p.errch <- true
	}()

	go wsutils.KeepAlive(p.conn, time.Second*15, p.errch)
	go p.WsReadRedis()
	go p.ReadDb()
	go p.WsWrite()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	p.Db.FuncApi(ctx, cancel, &models.Flowjson{Mode: "GetAllRooms"})

	for {
		flowjson := &models.Flowjson{}
		err := p.conn.ReadJSON(flowjson)
		if err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		go p.Db.FuncApi(ctx, cancel, flowjson)
	}
}

func (p *pubsub) WsReadRedis() {
	var err error
	chps := p.Pb.Channel()
	for {
		select {
		case action, ok := <-chps:
			if !ok {
				p.conn.Close()
				continue
			}
			flowjson := models.Flowjson{}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()
			if action.ErrorMsg == "unmarshall error" {
				p.Log.Error("unmarshalling error", err)
				p.conn.Close()
				return
			}
			switch {
			case contains(p.Db.PubSubActions(0), flowjson.Mode):
				err = p.Pb.Subscribe(ctx, fmt.Sprintf("%d%s", flowjson.Room, "room"))
			case contains(p.Db.PubSubActions(1), flowjson.Mode):
				err = p.Pb.Unsubscribe(ctx, fmt.Sprintf("%d%s", flowjson.Room, "room"))
			}
			if err != nil {
				p.Log.Error("publish || subscribe error", err)
				p.conn.Close()
				return
			}
			p.writeCh <- flowjson
		case <-p.errch:
			return
		}
	}
}

// stream for writing json to ws connection
func (p *pubsub) WsWrite() {
	for {
		select {
		case flowjson, ok := <-p.writeCh:
			if !ok {
				p.conn.Close()
				return
			}
			err := p.conn.WriteJSON(flowjson)
			if err != nil {
				p.Log.Error("writejson error", err)
				p.conn.Close()
				return
			}
		case <-p.errch:
			return
		}

	}

}
func (p *pubsub) ReadDb() {
	ch := p.Db.Channel()
	for {
		select {
		case flowjson, ok := <-ch:
			if !ok {
				p.conn.Close()
				continue
			}
			p.writeCh <- flowjson
			if len(flowjson.ErrorMsg) != 0 {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()
			payload, err := json.Marshal(&flowjson)
			if err != nil {
				p.Log.Error("unmarshalling error", err)
				p.conn.Close()
				return
			}
			switch {
			case contains(p.Db.PubSubActions(2), flowjson.Mode):
				err = p.Pb.Publish(ctx, fmt.Sprintf("%d%s", flowjson.Room, "room"), string(payload))
			default:
				i := 1
				for i = range flowjson.Users {
					err = p.Pb.Publish(ctx, fmt.Sprintf("%d%s", flowjson.Users[i], "user"), string(payload))
					if err != nil {
						break
					}
				}
			}
			if err != nil {
				p.Log.Error("publish error", err)
				p.conn.Close()
				return
			}
		case <-p.errch:
			return
		}
	}
}
func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
