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
	db   pubsub.Databaser
	pb   pubsub.PubSuber
	log  logger.Logger
}

func NewRouter(addr string, mp auth.ProviderMap, db pubsub.Databaser, pb pubsub.PubSuber, lg logger.Logger) *router {
	return &router{Addr: addr, Auth: mp, db: db, pb: pb, log: lg}
}
func (r *router) Run(ctx context.Context) error {

	middleware := middleware.UserMiddleware{Fetcher: r.Auth}
	pb := pubsub.PubSub{Db: r.db, Pb: r.pb}
	mux := http.NewServeMux()
	connect := http.HandlerFunc(pb.Connect)

	mux.HandleFunc("/beginauth", r.Auth.BeginAuth)
	mux.HandleFunc("/completeauth", r.Auth.CompleteAuth)
	mux.Handle("/connect", middleware.GetUserMW(connect))

	err := http.ListenAndServe(r.Addr, mux)
	if err != nil {
		return err
	}
	return nil
}
