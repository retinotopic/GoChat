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

func (p *PubSub) Connect(w http.ResponseWriter, r *http.Request) {
	sub := middleware.GetUser(r.Context())
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	p.conn = conn
	userid, err := p.Db.GetUserId(r.Context(), sub)
	p.UserId = userid
	p.WsHandle()
}

type Databaser interface {
	FuncApi(ctx context.Context, event *models.Event) error
	GetUserId(ctx context.Context, sub string) (uint32, error)
}

type Publisher interface {
	PublishWithSubscriptions(ctx context.Context, pubChannels []string, subChannel string, kind string) error
	Publish(ctx context.Context, channel string, message string) error
	Channel(ctx context.Context, closech <-chan bool, user string) <-chan []byte
}

// Publish||Subscribe Service
type PubSub struct {
	UserId uint32
	Pb     Publisher
	Db     Databaser
	Log    logger.Logger
	conn   *websocket.Conn
	errch  chan bool
}

// conn, err := upgrader.Upgrade(w, r, nil)

func (p *PubSub) WsHandle() {
	p.errch = make(chan bool, 10)
	defer func() {
		p.conn.CloseNow()
		p.errch <- true
	}()

	go p.ReadRedis()
	startevent := &models.Event{Event: "GetAllRooms"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	err := p.Db.FuncApi(ctx, startevent)
	if err != nil {
		p.conn.Close(websocket.StatusInternalError, "Database error, could not retrieve the initial data")
		return
	}
	for {
		_, b, err := p.conn.Read(context.TODO())
		if err != nil {
			return
		}
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()
			event := &models.Event{Data: b}
			event.GetEventName()
			err := p.Db.FuncApi(ctx, event)
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

func (p *PubSub) ReadRedis() {
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
