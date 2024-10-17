package router

import (
	"context"
	"net/http"

	"github.com/retinotopic/GoChat/internal/auth"
	"github.com/retinotopic/GoChat/internal/logger"
	"github.com/retinotopic/GoChat/internal/middleware"
	"github.com/retinotopic/GoChat/internal/pubsub"
)

type router struct {
	Addr string
	Auth auth.ProviderMap
	pb   pubsub.Publisher
	db   pubsub.Databaser
	log  logger.Logger
}

func NewRouter(addr string, mp auth.ProviderMap, pb pubsub.Publisher, db pubsub.Databaser, lg logger.Logger) *router {
	return &router{Addr: addr, Auth: mp, pb: pb, db: db, log: lg}
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

	mux.HandleFunc("/beginauth", r.Auth.BeginAuth)
	mux.HandleFunc("/completeauth", r.Auth.CompleteAuth)
	mux.Handle("/connect", middleware.GetUserMW(connect))

	return http.ListenAndServe(r.Addr, mux)
}
