package pubsub

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/goccy/go-json"

	"github.com/gorilla/websocket"
	"github.com/retinotopic/GoChat/internal/logger"
	"github.com/retinotopic/GoChat/internal/middleware"
	"github.com/retinotopic/GoChat/internal/models"
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
	go p.ReadDb()
	errs := make(chan error, 1)
	p.conn.SetPongHandler(func(string) error {
		p.conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		return nil
	})
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for {
			<-ticker.C
			err := p.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(9*time.Second))
			if err != nil {
				p.Log.Error("server ping error", err)
				errs <- err
			}
		}
	}()
	go p.WsReadRedis()

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
	chps := p.Db.Channel()
	for action := range chps {
		flowjson := models.Flowjson{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		if err = json.Unmarshal([]byte(action.Message), &flowjson); err != nil {
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
			err = p.Pb.Publish(ctx, fmt.Sprintf("%d%s", flowjson.Room, "room"), payload)
		default:
			i := 1
			for i = range flowjson.Users {
				err = p.Pb.Publish(ctx, fmt.Sprintf("%d%s", flowjson.Users[i], "user"), payload)
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
