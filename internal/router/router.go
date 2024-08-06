package router

import (
	"context"
	"net/http"
	"os"

	"github.com/retinotopic/GoChat/internal/auth"
	"github.com/retinotopic/GoChat/internal/db"
	"github.com/retinotopic/GoChat/internal/middleware"
	"github.com/retinotopic/GoChat/internal/pubsub"
)

type Router struct {
	Addr       string
	Auth       auth.ProviderMap
	PubSubAddr string
}

func NewRouter(addr string, mp auth.ProviderMap, addrpb string) *Router {
	return &Router{Addr: addr, Auth: mp, PubSubAddr: addrpb}
}
func (r *Router) Run() error {
	db, err := db.NewPool(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	middleware := middleware.UserMiddleware{GetProvider: r.Auth.GetProvider}

	mux := http.NewServeMux()
	pb := pubsub.NewPubsub(db, r.PubSubAddr)
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
