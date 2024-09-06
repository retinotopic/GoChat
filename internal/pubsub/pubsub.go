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
	p := pubsub{conn: conn, Pb: pb, Db: db, Log: c.Log}
	p.conn = conn
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
}

// conn, err := upgrader.Upgrade(w, r, nil)

func (p *pubsub) WsHandle() {
	defer func() {
		p.conn.Close()
	}()

	errch := make(chan error, 3)
	go wsutils.KeepAlive(p.conn, time.Second*15, errch)
	go p.WsReadRedis()
	go p.ReadDb()

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
	for action := range chps {
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
	}
}

// stream for writing json to ws connection
func (p *pubsub) WsWrite() {
	for flowjson := range p.writeCh {
		err := p.conn.WriteJSON(&flowjson)
		if err != nil {
			p.Log.Error("unmarshalling error", err)
			p.conn.Close()
			return
		}
	}
}
func (p *pubsub) ReadDb() {
	ch := p.Db.Channel()

	for flowjson := range ch {
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
