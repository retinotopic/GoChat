package pubsub

import (
	"context"
	"log"
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
	Channel(user string) <-chan []byte
}

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
	startevent := &models.EventMetadata{Event: "Get All Rooms", UserId: userid, Type: 2}
	p.ProcessEvent(startevent, conn)
	for {
		msgType, b, err := conn.Read(context.TODO()) // incoming client user requests
		if err != nil {
			log.Println("HUH?", err)
			return
		}
		if msgType == websocket.MessageText && len(b) > 0 {
			log.Println("new", err)

			event := &models.EventMetadata{}
			err := json.Unmarshal(b, &event)
			if err != nil {
				continue
			}
			event.UserId = userid
			log.Println(event)
			go p.ProcessEvent(event, conn)
		}
	}
}

func (p *PubSub) ReadPubSub(userid uint64, errch <-chan bool, conn *websocket.Conn) {
	userId := strconv.Itoa(int(userid))
	chps := p.Pb.Channel(userId)
	for {
		select {
		case b, ok := <-chps: // incoming successful events from others users
			if !ok {
				conn.CloseNow()
				return
			}
			err := WriteTimeout(time.Second*15, conn, b)
			if err != nil {
				conn.CloseNow()
			}
		case <-errch:
			return
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
	if err != nil {
		event.ErrorMsg = err.Error()
		event.OrderCmd[0] = 0
		log.Println("Process Event Func Api", err)
	}
	b, err := json.Marshal(event)
	if err != nil {
		conn.Close(websocket.StatusInternalError, "Marshal error")
	}

	for i := range event.OrderCmd {
		if event.OrderCmd[i] == 0 {
			err = WriteTimeout(time.Second*15, conn, b)
			if err != nil {
				conn.CloseNow()
			}
			break
		}
		switch event.OrderCmd[i] {
		case 1:
			err = p.Pb.PublishWithMessage(ctx, event.PublishChs, string(b))
			log.Println("publish with message")
			if err != nil {
				conn.Close(websocket.StatusInternalError, err.Error())
				p.Log.Error("Process Event: switch event.OrderCmd 1", err)
			}
		case 2:
			err = p.Pb.PublishWithSubscriptions(ctx, event.UserChs, event.PublishChs, event.Kind)
			log.Println("publish with sub")
			if err != nil {
				conn.Close(websocket.StatusInternalError, err.Error())
				p.Log.Error("Process Event: switch event.OrderCmd 2", err)
			}
		}
	}

	log.Println(string(b), "+", event.Event, event.UserId, event.ErrorMsg)

}
