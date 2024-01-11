package server

import (
	"GoChat/pkg/htmx"
	"fmt"
	"log"
	"net/http"
	"net/smtp"

	templ "github.com/a-h/templ"
)

type Server struct {
	addr         string
	port         string
	fromsmtp     string
	passwordsmtp string
	smtpHost     string
	smtpPort     string
	auth         smtp.Auth
}

func NewServer(addr string, fromsmtp string, passwordsmtp string, smtpHost string, smtpPort string) *Server {
	return &Server{addr: addr, fromsmtp: fromsmtp, passwordsmtp: passwordsmtp, smtpHost: smtpHost, smtpPort: smtpPort}
}
func (s *Server) Run() error {
	s.auth = smtp.PlainAuth("", s.fromsmtp, s.passwordsmtp, s.smtpHost)
	accountHandler := http.HandlerFunc(s.account)
	http.Handle("/main", templ.Handler(htmx.Main()))
	http.Handle("/login", templ.Handler(htmx.Login("")))
	http.Handle("/register", templ.Handler(htmx.Register()))
	http.Handle("/regnotif", s.MWregister(templ.Handler(htmx.Register_notification())))
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
func (s *Server) MWregister(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("im here", r.FormValue("email"))

		err := smtp.SendMail(s.smtpHost+":"+s.smtpPort, s.auth, s.fromsmtp, []string{r.FormValue("email")}, []byte("hiii"))
		if err != nil {
			fmt.Println(err)
			return
		}
		if r.FormValue("email") == "admin" && r.FormValue("password") == "ADMIN" {
			next.ServeHTTP(w, r)
		} else {
			log.Println(r.FormValue("email"), r.FormValue("password"), r.Form.Has("email"), r.Form.Has("password"))
			w.Write([]byte("not authorized"))
		}
	})
}
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("you are logged in"))
}
