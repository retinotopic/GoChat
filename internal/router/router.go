package router

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/retinotopic/GoChat/internal/auth"
	db "github.com/retinotopic/GoChat/internal/dbrepo/postgres"
	"github.com/retinotopic/GoChat/internal/middleware"
	"github.com/retinotopic/GoChat/internal/pubsub"
)

type router struct {
	Addr       string
	Auth       auth.ProviderMap
	PubSubAddr string
	DbAddr     string
}

func NewRouter(addr string, mp auth.ProviderMap, addrpb string, dbaddr string) *router {
	return &router{Addr: addr, Auth: mp, PubSubAddr: addrpb, DbAddr: dbaddr}
}
func (r *router) Run(ctx context.Context) error {
	pool, err := db.NewPool(ctx, r.DbAddr)
	if err != nil {
		return err
	}
	dbs := stdlib.OpenDBFromPool(pool.Pl)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(dbs, "migrations"); err != nil {
		return err
	}
	if err := dbs.Close(); err != nil {
		return err
	}
	middleware := middleware.UserMiddleware{FetchUser: r.Auth.FetchUser}
	pb := pubsub.PubSub{Db: &db.PgClient{}}
	mux := http.NewServeMux()
	connect := http.HandlerFunc(pb.Connect)

	mux.HandleFunc("/beginauth", r.Auth.BeginAuth)
	mux.HandleFunc("/completeauth", r.Auth.CompleteAuth)
	mux.Handle("/connect", middleware.GetUser(connect))

	err = http.ListenAndServe(r.Addr, mux)
	if err != nil {
		return err
	}
	return nil
}
