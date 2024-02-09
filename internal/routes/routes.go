package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/retinotopic/GoChat/internal/auth"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}
func (s *Server) Run() error {
	r := mux.NewRouter()
	r.HandleFunc("/{provider}/BeginLoginCreate", auth.BeginLoginCreateRoute)
	r.HandleFunc("/{provider}/CompleteLoginCreate", auth.CompleteLoginCreateRoute)
	r.HandleFunc("/refresh/{provider}", auth.RefreshRoute)
	r.HandleFunc("/refresh/revoke/{provider}", auth.RevokeRoute)
	return http.ListenAndServe(s.addr, r)

}
