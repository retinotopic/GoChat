package pubsub

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/goccy/go-json"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/retinotopic/GoChat/internal/db"
)

type userPubSub struct {
	db      *db.PgClient
	conn    *websocket.Conn
	writeCh chan *db.FlowJSON
	upubsub *redis.PubSub
	pub     *redis.Client
}

// conn, err := upgrader.Upgrade(w, r, nil)

func (u *userPubSub) WsHandle() {
	defer func() {
		u.conn.Close()
	}()
	go u.ReadDB()
	errs := make(chan error, 1)
	u.conn.SetPongHandler(func(string) error {
		u.conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		return nil
	})
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for {
			<-ticker.C
			err := u.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(9*time.Second))
			if err != nil {
				log.Println("server ping error:", err)
				errs <- err
			}
		}
	}()
	go u.WsReadRedis()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	u.db.FuncApi(ctx, cancel, &db.FlowJSON{Mode: "GetAllRooms"})

	for {
		flowjson := &db.FlowJSON{}
		err := u.conn.ReadJSON(flowjson)
		if err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		go u.db.FuncApi(ctx, cancel, flowjson)
	}
}
func (u *userPubSub) WsReadRedis() {
	var err error
	chps := u.upubsub.Channel()
	for message := range chps {
		flowjson := db.FlowJSON{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		if err = json.Unmarshal([]byte(message.Payload), &flowjson); err != nil {
			log.Println("unmarshalling error -> ", err)
			u.conn.Close()
			return
		}
		switch flowjson.Mode {
		case "CreateGroupRoom", "CreateDuoRoom", "AddUserToRoom":
			err = u.upubsub.Subscribe(ctx, fmt.Sprintf("%d%s", flowjson.Room, "room"))
		case "DeleteUsersFromRoom", "BlockUser":
			err = u.upubsub.Unsubscribe(ctx, fmt.Sprintf("%d%s", flowjson.Room, "room"))
		}
		if err != nil {
			log.Println(err, "publish || subscribe error ->", err)
			u.conn.Close()
			return
		}
		u.writeCh <- &flowjson
	}
}

// stream for writing json to ws connection
func (u *userPubSub) WsWrite() {
	for flowjson := range u.writeCh {
		err := u.conn.WriteJSON(&flowjson)
		if err != nil {
			log.Println("unmarshalling error ->", err)
			u.conn.Close()
			return
		}
	}
}
func (u *userPubSub) ReadDB() {
	ch := u.db.Channel()
	var intcmd *redis.IntCmd
	for flowjson := range ch {
		u.writeCh <- flowjson
		if flowjson.Err != nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		payload, err := json.Marshal(&flowjson)
		if err != nil {
			log.Println("unmarshalling error -> ", err)
			u.conn.Close()
			return
		}
		switch flowjson.Mode {
		case "SendMessage":
			intcmd = u.pub.Publish(ctx, fmt.Sprintf("%d%s", flowjson.Room, "room"), payload)
		default:
			i := 1
			for i = range flowjson.Users {
				if intcmd = u.pub.Publish(ctx, fmt.Sprintf("%d%s", flowjson.Users[i], "user"), payload); intcmd.Err() != nil {
					break
				}
			}
		}
		if intcmd.Err() != nil {
			log.Println("publish error ->", err)
			u.conn.Close()
			return
		}
	}
}
