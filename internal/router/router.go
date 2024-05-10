package router

import (
	"net/http"
)

type Router struct {
	addr string
}

func NewRouter(addr string) *Router {
	return &Router{addr: addr}
}
func (r *Router) Run() error {
	mux := http.NewServeMux()
	/*CurrentProviders := session.Session{
		"google":    google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), "http://localhost:8080/google/CompleteAuth"),
		"gfirebase": gfirebase.New(os.Getenv("WEB_API_KEY"), os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), "http://localhost:8080/gfirebase/CompleteAuth"),
	}
	FetchUser := http.HandlerFunc(CurrentProviders.FetchUser)
	mux.HandleFunc("/{provider}/BeginLoginCreate", CurrentProviders.BeginLoginCreate)
	mux.HandleFunc("/{provider}/CompleteLoginCreate", CurrentProviders.CompleteLoginCreate)
	mux.HandleFunc("/refresh/{provider}", CurrentProviders.Refresh)
	mux.HandleFunc("/refresh/revoke/{provider}", CurrentProviders.RevokeRefresh)
	mux.Handle("/chat", ws.WsConnect(FetchUser))*/
	return http.ListenAndServe(r.addr, mux)
}
