package server

import (
	"GoChat/pkg/htmx"
	"log"
	"net/http"

	templ "github.com/a-h/templ"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}
func (s *Server) Run() error {
	accounHandler := http.HandlerFunc(s.Account)
	http.Handle("/main", templ.Handler(htmx.Main()))
	http.Handle("/login", templ.Handler(htmx.Login("")))
	http.Handle("/register", templ.Handler(htmx.Register()))
	http.Handle("/account", s.mwAccount(accounHandler))
	return http.ListenAndServe(s.addr, nil)

}
func (s *Server) mwAccount(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("email") == "admin" && r.FormValue("password") == "ADMIN" {
			next.ServeHTTP(w, r)
		} else {
			log.Println(r.FormValue("email"), r.FormValue("password"), r.Form.Has("email"), r.Form.Has("password"))
			w.Write([]byte("not authorized"))
		}
	})
}

func (s *Server) Account(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("you are logged in"))
}
