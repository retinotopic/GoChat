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
	r.HandleFunc("/{provider}/CompleteAuth", func(w http.ResponseWriter, r *http.Request) {
		provider := mux.Vars(r)["provider"]
		auth.Providersmap[provider].CompleteAuth(w, r)
	})
	return http.ListenAndServe(s.addr, r)

}
