package server

import (
	"net/http"

	"github.com/retinotopic/GoChat/internal/auth/login"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}
func (s *Server) Run() error {
	http.HandleFunc("/login", login.LoginUser)

	return http.ListenAndServe(s.addr, nil)

}
