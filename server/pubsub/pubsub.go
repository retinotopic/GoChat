package pubsub

import (
	"context"
	"net/http"
	"strconv"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/coder/websocket"
	"github.com/retinotopic/GoChat/server/logger"
	"github.com/retinotopic/GoChat/server/middleware"
	"github.com/retinotopic/GoChat/server/models"
)

func (p *PubSub) Connect(w http.ResponseWriter, r *http.Request) {
	sub := middleware.GetUser(r.Context())
	b, userid, err := p.Db.GetUser(r.Context(), sub)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = WriteTimeout(time.Second*15, conn, b)
	if err != nil {
		conn.CloseNow()
	}
	p.WsHandle(userid, conn)
}

type Databaser interface {
	FuncApi(ctx context.Context, event *models.EventMetadata) error
	GetUser(ctx context.Context, sub string) ([]byte, uint64, error)
}

type Publisher interface {
	PublishWithSubscriptions(ctx context.Context, pubChannels []string, subChannels []string, kind string) error
	PublishWithMessage(ctx context.Context, channel []string, message string) error
	Channel(ctx context.Context, closech <-chan bool, user string) <-chan []byte
}

// Publish||Subscribe Service
type PubSub struct {
	Pb  Publisher
	Db  Databaser
	Log logger.Logger
}

func (p *PubSub) WsHandle(userid uint64, conn *websocket.Conn) {
	errch := make(chan bool, 10)
	defer func() {
		conn.CloseNow()
		errch <- true
	}()
	go p.ReadPubSub(userid, errch, conn)
	startevent := &models.EventMetadata{Event: "GetAllRooms"}
	p.ProcessEvent(startevent, conn)
	for {
		_, b, err := conn.Read(context.TODO()) // incoming client user requests
		if err != nil {
			return
		}
		event := &models.EventMetadata{Data: b}
		event.GetEventName()
		go p.ProcessEvent(event, conn)
	}
}

func (p *PubSub) ReadPubSub(userid uint64, errch <-chan bool, conn *websocket.Conn) {
	userId := strconv.Itoa(int(userid))
	closech := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	chps := p.Pb.Channel(ctx, closech, userId)
	for {
		select {
		case b := <-chps: // incoming successful events from others users
			err := WriteTimeout(time.Second*15, conn, b)
			if err != nil {
				conn.CloseNow()
			}
		case <-errch:
			closech <- true
		}
	}
}
func WriteTimeout(timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}
func (p *PubSub) ProcessEvent(event *models.EventMetadata, conn *websocket.Conn) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	err := p.Db.FuncApi(ctx, event)
	event.ErrorMsg = err.Error()
	for i := range event.OrderCmd {
		switch event.OrderCmd[i] {
		case 1:
			err = p.Pb.PublishWithMessage(ctx, event.SubForPub, string(event.Data))
			if err != nil {
				conn.Close(websocket.StatusInternalError, "PublishWithMessage error")
			}
		case 2:
			err = p.Pb.PublishWithSubscriptions(ctx, event.PubForSub, event.SubForPub, event.Kind)
			if err != nil {
				conn.Close(websocket.StatusInternalError, "PublishWithSubscriptions error")
			}
		}
	}

	bs, err := json.Marshal(event)
	if err != nil {
		conn.Close(websocket.StatusInternalError, "Marshal error")
	}
	err = WriteTimeout(time.Second*15, conn, bs)
	if err != nil {
		conn.CloseNow()
	}
}
