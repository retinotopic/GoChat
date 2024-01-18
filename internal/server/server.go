package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Server struct {
	addr string
}

var (
	config           *oauth2.Config
	oauthStateString = "sdghc34df6fg4ghgghdyjldprf-34t345234"
	authconf         *authconfig
)

type authconfig struct {
	Access_token  string `json:"access_token"`
	Expires_in    int    `json:"expires_in"`
	Id_token      string `json:"id_token"`
	Refresh_token string `json:"refresh_token"`
	Kid           string `json:"kid"`
}

func init() {
	config = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"openid email"},
		Endpoint:     google.Endpoint,
	}
	authconf = &authconfig{}
}
func NewServer(addr string) *Server {
	return &Server{addr: addr}
}
func (s *Server) Run() error {
	http.HandleFunc("/login", s.handleGoogleLogin)
	http.HandleFunc("/callback", s.handleGoogleCallback)
	http.HandleFunc("/callback", s.handleGoogleCallback)
	return http.ListenAndServe(s.addr, nil)

}
func (s *Server) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("im here1")
	url := config.AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline)
	fmt.Println(url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect) // composing our auth request url
}
func (s *Server) account(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	resp2, _ := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + token)
	err := resp2.Write(w)
	if err != nil {
		fmt.Println(err, "token id error")
	}
}
func (s *Server) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != oauthStateString {
		fmt.Println("invalid oauth state")
	}
	v := make(url.Values)
	v.Set("code", r.FormValue("code"))
	v.Set("client_id", config.ClientID)
	v.Set("client_secret", config.ClientSecret)
	v.Set("redirect_uri", config.RedirectURL)
	v.Set("grant_type", "authorization_code")
	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", strings.NewReader(v.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	json.NewDecoder(resp.Body).Decode(&authconf)
	resp.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(authconf.Access_token)
	fmt.Println(authconf.Expires_in)
	fmt.Println(authconf.Id_token)      // --------->>>>> to frontend httponly
	fmt.Println(authconf.Refresh_token) // --------->>>>> to frontend httponly
	resp2, _ := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + authconf.Id_token)
	err = resp2.Write(w)
}
