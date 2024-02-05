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
	r.HandleFunc("/{provider}/BeginAuth", auth.BeginAuthRoute)
	r.HandleFunc("/{provider}/CompleteAuth", auth.CompleteAuthRoute)
	r.HandleFunc("/refresh/{provider}", auth.RefreshRoute)
	r.HandleFunc("/refresh/revoke/{provider}", auth.RevokeRoute)
	return http.ListenAndServe(s.addr, r)

}
