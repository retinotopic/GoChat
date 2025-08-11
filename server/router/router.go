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
	db   pubsub.Databaser
	log  logger.Logger
}

func (rt *router) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func NewRouter(addr string, au middleware.Fetcher, db pubsub.Databaser, lg logger.Logger) *router {
	return &router{Addr: addr, Auth: au, db: db, log: lg}
}

func (rt *router) Run(ctx context.Context) error {

	middleware := middleware.UserMiddleware{Fetcher: rt.Auth}
	mux := http.NewServeMux()
	pubsub := pubsub.PubSub{
		Db:  rt.db,
		Log: rt.log,
	}
	pubsub.InitPS(5000)

	connect := http.HandlerFunc(pubsub.Connect)

	mux.Handle("/connect", middleware.GetUserMW(connect))
	mux.HandleFunc("/health", rt.HealthCheckHandler)

	switch os.Getenv("SSL_ENABLE") {
	case "true":
		return http.ListenAndServeTLS(rt.Addr, "/etc/ssl/certs/cert.pem", "/etc/ssl/private/key.pem", mux)
	default:
		return http.ListenAndServe(rt.Addr, mux)
	}
}
