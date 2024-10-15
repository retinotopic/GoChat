package pubsub

import (
	"context"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/retinotopic/GoChat/internal/logger"
	"github.com/retinotopic/GoChat/internal/middleware"
	"github.com/retinotopic/GoChat/internal/models"
)

type Connector struct {
	GetPubsub func(context.Context, uint32) (PubSuber, error)
	Db        Databaser
	Log       logger.Logger
}

func (c *Connector) Connect(w http.ResponseWriter, r *http.Request) {
	sub := middleware.GetUser(r.Context())
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		c.Log.Error("upgrade to websocket err", err)
		return
	}
	userid, err := c.Db.GetUser(r.Context(), sub)
	pb, err := c.GetPubsub(r.Context(), userid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	p := pubsub{conn: conn, Pb: pb, Db: c.Db, Log: c.Log, UserId: userid}
	p.WsHandle()
}

type Databaser interface {
	FuncApi(context.Context, context.CancelFunc, *models.Event) error
	GetUser(ctx context.Context, sub string) (uint32, error)
}
type PubSuber interface {
	PublishWithSubscriptions(ctx context.Context, pubChannels []string, subChannel string, kind string) error
	Publish(ctx context.Context, channel string, message string) error
	Channel(closech <-chan bool) <-chan []byte
}

// Publish||Subscribe Service
type pubsub struct {
	conn   *websocket.Conn
	UserId uint32
	Pb     PubSuber
	Db     Databaser
	Log    logger.Logger
	errch  chan bool
}

// conn, err := upgrader.Upgrade(w, r, nil)

func (p *pubsub) WsHandle() {
	p.errch = make(chan bool, 10)
	defer func() {
		p.conn.CloseNow()
		p.errch <- true
	}()

	go p.ReadRedis()
	startevent := &models.Event{Event: "GetAllRooms"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	err := p.Db.FuncApi(ctx, cancel, startevent)
	startevent.ErrorMsg = err.Error()
	WriteTimeout()
	for {
		_, b, err := p.conn.Read(context.TODO())
		if err != nil {
			return
		}
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()
			event := &models.Event{Data: b}
			err := p.Db.FuncApi(ctx, cancel, event)
			if len(event.SubChannel) != 0 {
				err = p.Pb.PublishWithSubscriptions(ctx, event.PubChannels, event.SubChannel, event.Kind)
				if err != nil {
					p.conn.Close(websocket.StatusInternalError, "internal error")
				}
			}
			if len(event.PubChannels) != 0 {
				p.Pb.Publish(ctx, event.PubChannels[0], string(event.Data))
				if err != nil {
					p.conn.Close(websocket.StatusInternalError, "internal error")
				}
			}
			WriteTimeout()
			p.conn.Write(ctx, 1, b)

		}()
	}
}

func (p *pubsub) ReadRedis() {
	var err error
	closech := make(chan bool, 1)
	chps := p.Pb.Channel(closech)
	for {
		select {
		case b := <-chps:
			err := p.conn.Write(ctx, 1, b)
		case <-p.errch:
			closech <- true
		}
	}
}
func WriteTimeout(timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}
