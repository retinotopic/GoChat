package router

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/retinotopic/GoChat/internal/auth"
	"github.com/retinotopic/GoChat/internal/db"
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
	db := stdlib.OpenDBFromPool(pool.Pl)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}
	if err := db.Close(); err != nil {
		return err
	}
	middleware := middleware.UserMiddleware{GetProvider: r.Auth.GetProvider}

	mux := http.NewServeMux()
	pb := pubsub.NewPubsub(pool, r.PubSubAddr)
	connect := http.HandlerFunc(pb.Connect)

	mux.HandleFunc("/beginauth", r.Auth.BeginAuth)
	mux.HandleFunc("/completeauth", r.Auth.CompleteAuth)
	mux.Handle("/connect", middleware.FetchUser(connect))

	err = http.ListenAndServe(r.Addr, mux)
	if err != nil {
		return err
	}
	return nil
}
