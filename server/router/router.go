package router

import (
	"context"
	"net/http"
	"os"

	"github.com/retinotopic/GoChat/server/logger"
	"github.com/retinotopic/GoChat/server/middleware"
	"github.com/retinotopic/GoChat/server/pubsub"
)

type router struct {
	Addr string
	Auth middleware.Fetcher
	pb   pubsub.Publisher
	db   pubsub.Databaser
	log  logger.Logger
}

func NewRouter(addr string, au middleware.Fetcher, pb pubsub.Publisher, db pubsub.Databaser, lg logger.Logger) *router {
	return &router{Addr: addr, Auth: au, pb: pb, db: db, log: lg}
}
func (r *router) Run(ctx context.Context) error {

	middleware := middleware.UserMiddleware{Fetcher: r.Auth}
	mux := http.NewServeMux()
	pubsub := pubsub.PubSub{
		Db:  r.db,
		Pb:  r.pb,
		Log: r.log,
	}
	connect := http.HandlerFunc(pubsub.Connect)

	mux.Handle("/connect", middleware.GetUserMW(connect))
	switch os.Getenv("SSL_ENABLE") {
	case "true":
		return http.ListenAndServeTLS(r.Addr, "/etc/ssl/certs/cert.pem", "/etc/ssl/private/key.pem", mux)
	default:
		return http.ListenAndServe(r.Addr, mux)
	}
}
