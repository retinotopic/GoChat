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
	accountHandler := http.HandlerFunc(s.account)
	http.Handle("/main", templ.Handler(htmx.Main()))
	http.Handle("/login", templ.Handler(htmx.Login("")))
	http.Handle("/register", templ.Handler(htmx.Register()))
	http.Handle("/account", s.MWaccount(accountHandler))
	return http.ListenAndServe(s.addr, nil)

}
func (s *Server) MWaccount(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("email") == "admin" && r.FormValue("password") == "ADMIN" {
			next.ServeHTTP(w, r)
		} else {
			log.Println(r.FormValue("email"), r.FormValue("password"), r.Form.Has("email"), r.Form.Has("password"))
			w.Write([]byte("not authorized"))
		}
	})
}

func (s *Server) account(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("you are logged in"))
}
func (s *Server) MWregister(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("you are logged in"))
}
