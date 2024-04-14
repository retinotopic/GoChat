package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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
		err = wsc.WsHandle()
		if err != nil {
			//write error to plain http
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
func (h HandlerWS) WsHandle() error {
	defer func() {
		h.DBc.Conn.Release()
		h.Conn.Close()
	}()
	go h.ReadDB()
	h.DBc.GetAllRooms(&db.FlowJSON{})
	go h.WsReadRedis()
	go h.WsReceiveClient()
	return nil
}
func (h *HandlerWS) WsWriteRedis() {
	for {
		flowjson := <-h.jsonCh
		payload, err := json.Marshal(&flowjson)
		if err != nil {
			log.Fatalln(err, "unmarshalling error")
		}
		switch flowjson.Mode {
		case "SendMessage":
			redisClient.Publish(context.Background(), fmt.Sprintf("%d %s", flowjson.Rooms[0], "room"), payload)
		case "CreateDuoRoom":
			redisClient.Publish(context.Background(), fmt.Sprintf("%d %s", flowjson.Users[1], "room"), payload)
		case "CreateGroupRoom":
			for i := range flowjson.Users {
				redisClient.Publish(context.Background(), fmt.Sprintf("%d %s", flowjson.Users[i], "user"), payload)
			}
		}
	}
}
func (h *HandlerWS) WsReadRedis() {
	rps := redisClient.Subscribe(context.Background(), "chat")
	flowjson := db.FlowJSON{}
	for {
		message, err := rps.ReceiveMessage(context.Background())
		if err != nil {
			log.Println(err)
		}
		if err := json.Unmarshal([]byte(message.Payload), &flowjson); err != nil {
			log.Fatalln(err, "unmarshalling error")
		}
		h.WriteCh <- flowjson
	}
}
func (h *HandlerWS) WsReceiveClient() {
	for {
		flowjson := &db.FlowJSON{}
		err := h.Conn.ReadJSON(flowjson)
		if err != nil {
			return
		}
		go h.DBc.TxManage(flowjson)
	}
}
func (h *HandlerWS) WsWrite() {
	for {
		flowjson := <-h.WriteCh
		err := h.Conn.WriteJSON(&flowjson)
		if err != nil {
			return
		}
	}
}
func (h *HandlerWS) ReadDB() {
	for {
		flowjson := h.DBc.ReadFlowjson()
		h.jsonCh <- flowjson
	}
}
