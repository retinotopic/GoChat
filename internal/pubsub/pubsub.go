package pubsub

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/retinotopic/GoChat/internal/db"
	"github.com/retinotopic/GoChat/internal/middleware"
)

func NewPubsub(db *db.Pool, addr string) *pubsub {
	redisClient := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	u := &pubsub{
		Pool:   db,
		Client: redisClient,
	}
	return u
}

type pubsub struct {
	*db.Pool
	*redis.Client
}

func (p *pubsub) newUserPubSub(ctx context.Context, dbc *db.PgClient, conn *websocket.Conn) *userPubSub {
	u := &userPubSub{
		db:      dbc,
		conn:    conn,
		writeCh: make(chan *db.FlowJSON, 100),
		upubsub: p.Subscribe(ctx, fmt.Sprintf("%d%s", dbc.UserID, "user")),
		pub:     p.Client,
	}
	return u
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (p *pubsub) Connect(w http.ResponseWriter, r *http.Request) {
	sub := middleware.GetUser(r.Context())
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	db, err := p.NewClient(r.Context(), sub)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	wsc := p.newUserPubSub(r.Context(), db, conn)
	wsc.WsHandle()
}
