package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/retinotopic/GoChat/internal/auth/auth"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}
func (s *Server) Run() error {
	r := mux.NewRouter()
	r.HandleFunc("/{provider}/BeginAuth", auth.BeginAuthenticator)
	r.HandleFunc("/{provider}/CompleteAuth", auth.CompleteAuthenticator)
	return http.ListenAndServe(s.addr, r)

}
