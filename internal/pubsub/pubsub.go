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

func NewPubsub(db *db.Pool, addr string) *Pubsub {
	redisClient := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	u := &Pubsub{
		Pool:   db,
		Client: redisClient,
	}
	return u
}

type Pubsub struct {
	*db.Pool
	*redis.Client
}

func (p *Pubsub) newUserPubSub(ctx context.Context, dbc *db.PgClient, conn *websocket.Conn) *UserPubSub {
	u := &UserPubSub{
		Db:      dbc,
		Conn:    conn,
		WriteCh: make(chan db.FlowJSON, 100),
		Upubsub: p.Subscribe(ctx, fmt.Sprintf("%d%s", dbc.UserID, "user")),
	}
	return u
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (p *Pubsub) Connect(w http.ResponseWriter, r *http.Request) {
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
