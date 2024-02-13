package server

import (
	"net/http"

	"github.com/retinotopic/GoChat/internal/auth"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}
func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/{provider}/BeginLoginCreate", func(w http.ResponseWriter, r *http.Request) {
		auth.RefreshI(auth.CurrentProviders[r.PathValue("provider")]).ServeHTTP(w, r)
	})
	mux.HandleFunc("/{provider}/CompleteLoginCreate", func(w http.ResponseWriter, r *http.Request) {
		auth.CompleteLoginCreateI(auth.CurrentProviders[r.PathValue("provider")]).ServeHTTP(w, r)
	})
	mux.HandleFunc("/refresh/{provider}", func(w http.ResponseWriter, r *http.Request) {
		auth.RefreshI(auth.CurrentProviders[r.PathValue("provider")]).ServeHTTP(w, r)
	})
	mux.HandleFunc("/refresh/revoke/{provider}", func(w http.ResponseWriter, r *http.Request) {
		auth.RevokeI(auth.CurrentProviders[r.PathValue("provider")]).ServeHTTP(w, r)
	})
	return http.ListenAndServe(s.addr, mux)
}
