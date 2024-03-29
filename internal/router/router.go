package router

import (
	"net/http"
	"os"

	"github.com/retinotopic/GoChat/internal/api/ws"
	"github.com/retinotopic/GoChat/internal/db"
	"github.com/retinotopic/GoChat/internal/providers/gfirebase"
	"github.com/retinotopic/GoChat/internal/providers/google"
	"github.com/retinotopic/GoChat/internal/session"
)

type Router struct {
	addr string
}

func NewRouter(addr string) *Router {
	return &Router{addr: addr}
}
func (r *Router) Run() error {
	conn, err := db.ConnectToDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	wsh := ws.NewHandlerWS(conn)
	CurrentProviders := session.Session{
		"google":    google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), "http://localhost:8080/google/CompleteAuth"),
		"gfirebase": gfirebase.New(os.Getenv("WEB_API_KEY"), os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), "http://localhost:8080/gfirebase/CompleteAuth"),
	}
	FetchUser := http.HandlerFunc(CurrentProviders.Fetcher)
	mux := http.NewServeMux()
	mux.HandleFunc("/{provider}/BeginLoginCreate", CurrentProviders.BeginLoginCreate)
	mux.HandleFunc("/{provider}/CompleteLoginCreate", CurrentProviders.CompleteLoginCreate)
	mux.HandleFunc("/refresh/{provider}", CurrentProviders.Refresh)
	mux.HandleFunc("/refresh/revoke/{provider}", CurrentProviders.Revoke)
	mux.Handle("/chat", wsh.WsConnect(FetchUser))
	return http.ListenAndServe(r.addr, mux)
}
