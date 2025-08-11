package pubsub

import (
	"context"
	"errors"
	"github.com/puzpuzpuz/xsync/v4"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/coder/websocket"
	"github.com/retinotopic/GoChat/server/logger"
	"github.com/retinotopic/GoChat/server/middleware"
	"github.com/retinotopic/GoChat/server/models"
	"github.com/valkey-io/valkey-go/valkeylimiter"
)

func (p *PubSub) Connect(w http.ResponseWriter, r *http.Request) {
	sub := middleware.GetUser(r.Context())
	b, userid, err := p.Db.GetUser(r.Context(), sub)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	p.Log.Error(r.RemoteAddr, errors.New("CHECK THIS"))
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

type PubSub struct {
	Db        Databaser
	RoomsPool sync.Pool
	ChanPool  sync.Pool
	Rooms     *xsync.Map[string, *Room]
	Users     *xsync.Map[string, *User]
	Limiter   valkeylimiter.RateLimiterClient
	IsDebug   bool
	Log       logger.Logger
}

func (p *PubSub) WsHandle(userid uint64, conn *websocket.Conn) {
	errch := make(chan bool, 10)
	defer func() {
		conn.CloseNow()
		errch <- true
	}()
	userId := strconv.Itoa(int(userid))
	ch := p.GetUser(userId)
	if ch == nil {
		return
	}
	defer p.ReleaseUser(userId, ch)

	startevent := &models.EventMetadata{Event: "Get All Rooms", UserId: userid, Type: 2}
	p.ProcessEvent(startevent, conn)
	go p.ReadPubSub(conn, userid, ch, errch)
	for {
		msgType, b, err := conn.Read(context.TODO()) // incoming client user requests
		if err != nil {
			return
		}
		if msgType == websocket.MessageText && len(b) > 0 {
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

func (p *PubSub) ReadPubSub(conn *websocket.Conn, userid uint64, ch <-chan []byte, errch <-chan bool) {
	for {
		select {
		case b := <-ch: // incoming successful events from others users
			p.Log.Error(string(b), errors.New("CHECK THIS"))
			err := WriteTimeout(time.Second*15, conn, b)
			if err != nil {
				conn.CloseNow()
			}
		case <-errch:
			p.Log.Error("errch", errors.New("CHECK THIS"))
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
	if event.OrderCmd[0] == 0 {
		err = WriteTimeout(time.Second*15, conn, b)
		if err != nil {
			conn.CloseNow()
		}
	}

	for i := range event.OrderCmd {
		switch event.OrderCmd[i] {
		case 1:
			p.PublishWithMessage(event.PublishChs[0], b)
		case 2:
			p.SubscribeUsers(event.UserChs, event.PublishChs[0], event.Kind, false)
		case 3:
			rm := []models.RoomClient{}
			err := json.Unmarshal(event.Data, &rm)
			if err != nil {
				conn.CloseNow()
				p.Log.Error("Unmarshal initial subcribe", err)
				return
			}
			for j := range rm {
				users := []string{}
				rmid := strconv.FormatUint(rm[j].RoomId, 10)
				for k := range rm[j].Users {
					users = append(users, strconv.FormatUint(rm[j].Users[k].UserId, 10))
				}
				p.SubscribeUsers(users, rmid, true, true)
			}

		}
	}

	log.Println(string(b), "+", event.Event, event.UserId, event.ErrorMsg)
}
