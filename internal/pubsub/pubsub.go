package pubsub

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/goccy/go-json"

	"github.com/gorilla/websocket"
	"github.com/retinotopic/GoChat/internal/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Databaser interface {
	FuncApi(context.Context, context.CancelFunc, *models.Flowjson)
	PubSubActions(int) [][]string
	Channel() <-chan models.Flowjson
}
type PubSuber interface {
	Unsubscribe(context.Context, ...string) error
	Subscribe(context.Context, ...string) error
	Publish(context.Context, string, interface{}) error
}

type PubSub struct {
	conn    *websocket.Conn
	writeCh chan models.Flowjson
	pb      PubSuber
	db      Databaser
}

// conn, err := upgrader.Upgrade(w, r, nil)

func (p *PubSub) WsHandle() {
	defer func() {
		p.conn.Close()
	}()
	go p.ReadDB()
	errs := make(chan error, 1)
	p.conn.SetPongHandler(func(string) error {
		p.conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		return nil
	})
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for {
			<-ticker.C
			err := p.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(9*time.Second))
			if err != nil {
				log.Println("server ping error:", err)
				errs <- err
			}
		}
	}()
	go p.WsReadRedis()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	p.db.FuncApi(ctx, cancel, &models.Flowjson{Mode: "GetAllRooms"})

	for {
		flowjson := &models.Flowjson{}
		err := p.conn.ReadJSON(flowjson)
		if err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		go p.db.FuncApi(ctx, cancel, flowjson)
	}
}
func (p *PubSub) WsReadRedis() {
	var err error
	chps := p.db.Channel()
	for action := range chps {
		flowjson := models.Flowjson{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		if err = json.Unmarshal([]byte(action.Message), &flowjson); err != nil {
			log.Println("unmarshalling error -> ", err)
			p.conn.Close()
			return
		}
		switch flowjson.Mode {
		case "CreateGroupRoom", "CreateDuoRoom", "AddUserToRoom":
			err = p.pb.Subscribe(ctx, fmt.Sprintf("%d%s", flowjson.Room, "room"))
		case "DeleteUsersFromRoom", "BlockUser":
			err = p.pb.Unsubscribe(ctx, fmt.Sprintf("%d%s", flowjson.Room, "room"))
		}
		if err != nil {
			log.Println(err, "publish || subscribe error ->", err)
			p.conn.Close()
			return
		}
		p.writeCh <- flowjson
	}
}

// stream for writing json to ws connection
func (p *PubSub) WsWrite() {
	for flowjson := range p.writeCh {
		err := p.conn.WriteJSON(&flowjson)
		if err != nil {
			log.Println("unmarshalling error ->", err)
			p.conn.Close()
			return
		}
	}
}
func (p *PubSub) ReadDB() {
	ch := p.db.Channel()

	for flowjson := range ch {
		p.writeCh <- flowjson
		if len(flowjson.ErrorMsg) != 0 {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		payload, err := json.Marshal(&flowjson)
		if err != nil {
			log.Println("unmarshalling error -> ", err)
			p.conn.Close()
			return
		}
		switch flowjson.Mode {
		case "SendMessage":
			err = p.pb.Publish(ctx, fmt.Sprintf("%d%s", flowjson.Room, "room"), payload)
		default:
			i := 1
			for i = range flowjson.Users {
				err = p.pb.Publish(ctx, fmt.Sprintf("%d%s", flowjson.Users[i], "user"), payload)
				if err != nil {
					break
				}
			}
		}
		if err != nil {
			log.Println("publish error ->", err)
			p.conn.Close()
			return
		}
	}
}
