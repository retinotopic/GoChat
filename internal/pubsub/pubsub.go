package pubsub

import (
	"context"
	"net/http"
	"strconv"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/coder/websocket"
	"github.com/retinotopic/GoChat/internal/logger"
	"github.com/retinotopic/GoChat/internal/middleware"
	"github.com/retinotopic/GoChat/internal/models"
)

func (p *PubSub) Connect(w http.ResponseWriter, r *http.Request) {
	sub := middleware.GetUser(r.Context())
	userid, username, err := p.Db.GetUserId(r.Context(), sub)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.Write([]byte(`{"UserId":` + strconv.Itoa(int(userid)) + `,"Username":"` + username + `"}`))

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	p.conn = conn
	p.UserId = userid
	p.WsHandle()
}

type Databaser interface {
	FuncApi(ctx context.Context, event *models.Event) error
	GetUserId(ctx context.Context, sub string) (uint32, string, error)
}

type Publisher interface {
	PublishWithSubscriptions(ctx context.Context, pubChannels []string, subChannels []string, kind string) error
	PublishWithMessage(ctx context.Context, channel []string, message string) error
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

func (p *PubSub) WsHandle() {
	p.errch = make(chan bool, 10)
	defer func() {
		p.conn.CloseNow()
		p.errch <- true
	}()
	go p.ReadRedis()
	startevent := &models.Event{Event: "GetAllRooms"}
	p.ProcessEvent(startevent)
	for {
		_, b, err := p.conn.Read(context.TODO())
		if err != nil {
			return
		}
		event := &models.Event{Data: b}
		event.GetEventName()
		p.ProcessEvent(event)
	}
}

func (p *PubSub) ReadRedis() {
	closech := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	chps := p.Pb.Channel(ctx, closech, strconv.Itoa(int(p.UserId)))
	for {
		select {
		case b := <-chps:
			err := WriteTimeout(time.Second*15, p.conn, b)
			if err != nil {
				p.conn.CloseNow()
			}
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
func (p *PubSub) ProcessEvent(event *models.Event) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	err := p.Db.FuncApi(ctx, event)

	event.ErrorMsg = err.Error()
	for i := range event.OrderCmd {
		switch event.OrderCmd[i] {
		case 1:
			err = p.Pb.PublishWithMessage(ctx, event.SubForPub, string(event.Data))
			if err != nil {
				p.conn.Close(websocket.StatusInternalError, "internal error")
			}
		case 2:
			err = p.Pb.PublishWithSubscriptions(ctx, event.PubForSub, event.SubForPub, event.Kind)
			if err != nil {
				p.conn.Close(websocket.StatusInternalError, "internal error")
			}
		}
	}

	bs, err := json.Marshal(event)
	if err != nil {
		p.conn.Close(websocket.StatusInternalError, "internal error")
	}
	WriteTimeout(time.Second*15, p.conn, bs)

}
