package ws

import (
	"context"
	"encoding/json"
	"fmt"
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
		dbClient, err := db.NewClient(sub, h.db)
		if err != nil {
			log.Println("wrong sub", err)
			return
		}
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
	FuncMap := make(map[string]func(*db.FlowJSON))
	FuncMap["SendMessage"] = dbc.SendMessage
	FuncMap["GetMessages"] = dbc.GetMessages
	FuncMap["CreateDuoRoom"] = dbc.CreateDuoRoom
	FuncMap["CreateGroupRoom"] = dbc.CreateRoom
	go h.WsReceive(FuncMap, conn, dbc)
	go h.WsSend(FuncMap, conn, dbc)
}
func (h HandlerWS) WsReceive(funcMap map[string]func(*db.FlowJSON), conn *websocket.Conn, dbc *db.PostgresClient) {

	rps := h.rdb.Subscribe(context.Background(), "chat")
	flowjson := &db.FlowJSON{}
	for {
		message, err := rps.ReceiveMessage(context.Background())
		if err != nil {
			log.Println(err)
		}
		if err := json.Unmarshal([]byte(message.Payload), &flowjson); err != nil {
			log.Fatalln(err, "unmarshalling error")
		}
		dbc.UpdateRealtimeInfo(flowjson)
		if err != nil {
			log.Println(err)
		}
		err = conn.WriteJSON(flowjson)
		if err != nil {
			log.Println(err)
			break
		}
	}
}
func (h HandlerWS) WsSend(funcMap map[string]func(*db.FlowJSON), conn *websocket.Conn, dbc *db.PostgresClient) {
	for {
		flowjson := &db.FlowJSON{}
		err := conn.ReadJSON(flowjson)
		if err != nil {
			log.Println(err)
			break
		}
		dbc.TxBegin(flowjson)
		if flowjson.Err != nil {
			funcMap[flowjson.Mode](flowjson)
		}
		dbc.TxCommit(flowjson)
		if flowjson.Err != nil {
			return
		}

		if flowjson.Status == "bad" || flowjson.Status == "senderonly" {
			conn.WriteJSON(flowjson)
		} else {
			dbc.UpdateRealtimeInfo(flowjson)
			payload, err := json.Marshal(flowjson)
			if err != nil {
				log.Fatalln(err, "marshalling error")
			}
			if err := h.rdb.Publish(context.Background(), fmt.Sprintf("%d", flowjson.Room), payload).Err(); err != nil {
				log.Println(err)
			}
		}
	}
}
