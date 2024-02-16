package router

import (
	"net/http"
	"os"

	"github.com/retinotopic/GoChat/internal/providers/gfirebase"
	"github.com/retinotopic/GoChat/internal/providers/google"
	session "github.com/retinotopic/GoChat/internal/session"
)

type Router struct {
	addr string
}

func NewRouter(addr string) *Router {
	return &Router{addr: addr}
}
func (s *Router) Run() error {
	CurrentProviders := session.Session{
		"google":    google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), "http://localhost:8080/google/CompleteAuth"),
		"gfirebase": gfirebase.New(os.Getenv("WEB_API_KEY"), os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), "http://localhost:8080/gfirebase/CompleteAuth"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/{provider}/BeginLoginCreate", CurrentProviders.BeginLoginCreate)
	mux.HandleFunc("/{provider}/CompleteLoginCreate", CurrentProviders.CompleteLoginCreate)
	mux.HandleFunc("/refresh/{provider}", CurrentProviders.Refresh)
	mux.HandleFunc("/refresh/revoke/{provider}", CurrentProviders.Revoke)
	return http.ListenAndServe(s.addr, mux)
}
