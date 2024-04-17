package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/retinotopic/GoChat/internal/db"
	"github.com/retinotopic/GoChat/pkg/safectx"
)

type HandlerWS struct {
	DBc     *db.PostgresClient
	Conn    *websocket.Conn
	WriteCh chan db.FlowJSON
	jsonCh  chan db.FlowJSON
	pubsub  *redis.PubSub
}

// conn, err := upgrader.Upgrade(w, r, nil)
func NewHandlerWS(dbc *db.PostgresClient, conn *websocket.Conn) *HandlerWS {
	return &HandlerWS{
		DBc:     dbc,
		Conn:    conn,
		WriteCh: make(chan db.FlowJSON, 10),
		jsonCh:  make(chan db.FlowJSON, 10),
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func WsConnect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		sub, ok := safectx.GetContextString(r.Context(), "sub")
		if !ok {
			log.Println("no sub")
			return
		}
		dbconn, err := db.ConnectToDB(os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Println("wrong sub", err)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		dbc, err := db.NewClient(sub, dbconn)
		if err != nil {
			//write error to plain http
			w.WriteHeader(http.StatusInternalServerError)
		}
		wsc := NewHandlerWS(dbc, conn)
		wsc.WsHandle()
	})
}
func (h HandlerWS) WsHandle() {
	defer func() {
		h.DBc.Conn.Release()
		h.Conn.Close()
	}()
	go h.ReadDB()
	errs := make(chan error, 1)
	h.Conn.SetPongHandler(func(string) error {
		h.Conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		return nil
	})
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for {
			<-ticker.C
			err := h.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(9*time.Second))
			if err != nil {
				log.Println("server ping error:", err)
				errs <- err
			}
		}
	}()
	h.pubsub = redisClient.Subscribe(context.Background(), fmt.Sprintf("%d%s", h.DBc.UserID, "user"))
	h.DBc.GetAllRooms(&db.FlowJSON{})
	go h.WsReadRedis()
	for {
		flowjson := &db.FlowJSON{}
		err := h.Conn.ReadJSON(flowjson)
		if err != nil {
			return
		}
		go h.DBc.TxManage(flowjson)
	}
}
func (h *HandlerWS) WsReadRedis() {
	flowjson := db.FlowJSON{}
	for {
		message, err := h.pubsub.ReceiveMessage(context.Background())
		if err != nil {
			continue
		}
		if err := json.Unmarshal([]byte(message.Payload), &flowjson); err != nil {
			log.Fatalln(err, "unmarshalling error")
		}
		switch flowjson.Mode {
		case "SendMessage":
			h.WriteCh <- flowjson
		case "CreateGroupRoom", "CreateDuoRoom", "AddUserToRoom":
			h.pubsub.Subscribe(context.Background(), fmt.Sprintf("%d%s", flowjson.Rooms[0], "room"))
		case "DeleteUsersFromRoom", "BlockUser":
			h.pubsub.Unsubscribe(context.Background(), fmt.Sprintf("%d%s", flowjson.Rooms[0], "room"))
		}
		h.WriteCh <- flowjson
	}
}

// stream for writing json to ws connection
func (h *HandlerWS) WsWrite() {
	for flowjson := range h.WriteCh {
		err := h.Conn.WriteJSON(&flowjson)
		if err != nil {
			return
		}
	}
}
func (h *HandlerWS) ReadDB() {
	for {
		flowjson := h.DBc.ReadFlowjson()
		h.WriteCh <- flowjson
		if flowjson.Err != nil {
			continue
		}
		payload, err := json.Marshal(&flowjson)
		if err != nil {
			log.Fatalln(err, "unmarshalling error")
		}
		switch flowjson.Mode {
		case "SendMessage":
			redisClient.Publish(context.Background(), fmt.Sprintf("%d%s", flowjson.Rooms[0], "room"), payload)
		default:
			i := 1
			for i = range flowjson.Users {
				redisClient.Publish(context.Background(), fmt.Sprintf("%d%s", flowjson.Users[i], "user"), payload)
			}
		}
	}
}
