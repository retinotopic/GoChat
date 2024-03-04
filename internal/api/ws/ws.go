package ws

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/retinotopic/GoChat/internal/db"
	"github.com/retinotopic/GoChat/pkg/safectx"
)

type HandlerWS struct {
	upgrader *websocket.Upgrader
	db       *pgxpool.Pool
	rdb      *redis.Client
}
type tempJSON struct {
	Mode string `json:"Mode"`
	Data string `json:"Data"`
}

// conn, err := upgrader.Upgrade(w, r, nil)
func NewHandlerWS(dbc *pgxpool.Pool) *HandlerWS {
	return &HandlerWS{
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		db: dbc,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h HandlerWS) WsConnect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		sub, ok := safectx.GetContextString(r.Context(), "sub")
		if !ok {
			log.Println("no sub")
			return
		}
		///////

		// "AI" SELECT ALL MESSAGES FROM ROOMS

		///////
		dbClient, err := db.NewClient(sub, h.db)
		if err != nil {
			log.Println("wrong sub", err)
			return
		}
		dbClient.GetMessages(0, 0)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		if err != nil {
			log.Println(err)
			return
		}
		h.WsHandle(dbClient, conn)
	})
}
func (h HandlerWS) WsHandle(dbc *db.PostgresClient, conn *websocket.Conn) {
	defer func() {
		dbc.Conn.Release()
		conn.Close()
	}()
	FuncMap := make(map[string]func(string) error)
	FuncMap["SendMessage"] = dbc.SendMessage
	FuncMap["CreateDuoRoom"] = dbc.CreateDuoRoom
	FuncMap["CreateGroupRoom"] = dbc.CreateGroupRoom
	go h.WsReceive(FuncMap, conn)
	go h.WsSend(FuncMap, conn)
}
func (h HandlerWS) WsReceive(funcMap map[string]func(string) error, conn *websocket.Conn) {

	rps := h.rdb.Subscribe(context.Background(), "chat")
	for {
		message, err := rps.ReceiveMessage(context.Background())
		message.Pattern = "chat"
		if err != nil {
			log.Println(err)
			break
		}
		err = conn.WriteJSON(message.Payload)
		if err != nil {
			log.Println(err)
			break
		}
	}
}
func (h HandlerWS) WsSend(funcMap map[string]func(string) error, conn *websocket.Conn) {
	tempjson := tempJSON{}
	for {
		err := conn.ReadJSON(tempjson)
		if err != nil {
			log.Println(err)
			break
		}
		err = funcMap[tempjson.Mode](tempjson.Data)
		if err != nil {
			log.Println(err)
			break
		}
		err = h.rdb.Publish(context.Background(), "chat", "placeholdfer message").Err()
		if err != nil {
			log.Println(err)
			break
		}
	}
}
