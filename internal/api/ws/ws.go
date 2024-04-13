package ws

import (
	"context"
	"encoding/json"
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
	WriteCh chan []byte
}

// conn, err := upgrader.Upgrade(w, r, nil)
func NewHandlerWS(dbc *db.PostgresClient, conn *websocket.Conn) *HandlerWS {
	return &HandlerWS{
		DBc:     dbc,
		Conn:    conn,
		WriteCh: make(chan []byte, 256),
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
	h.DBc.GetAllRooms(&db.FlowJSON{})

	go h.WsReceiveRedis()
	go h.WsReceiveClient()
	return nil
}
func (h *HandlerWS) WsReceiveRedis() {
	rps := redisClient.Subscribe(context.Background(), "chat")
	flowjson := &db.FlowJSON{}

	for {
		message, err := rps.ReceiveMessage(context.Background())
		if err != nil {
			log.Println(err)
		}
		if err := json.Unmarshal([]byte(message.Payload), &flowjson); err != nil {
			log.Fatalln(err, "unmarshalling error")
		}
		if err != nil {
			log.Println(err)
		}
		err = h.Conn.WriteJSON(flowjson)
		if err != nil {
			log.Println(err)
			break
		}
	}
}
func (h *HandlerWS) WsReceiveClient() {
	for {
		flowjson := &db.FlowJSON{}
		err := h.Conn.ReadJSON(flowjson)
		if err != nil {
			return
		}
		h.DBc.TxManage(flowjson)
	}
}
func (h *HandlerWS) WsWrite() {
	for {
		message := <-h.WriteCh
		err := h.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}
func (h *HandlerWS) ReadDB() {
	for {
		h.DBc.ReadFlowjson()
	}
}
