package gfirebase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/retinotopic/pokerGO/pkg/randfuncs"
	"google.golang.org/api/option"
)

type Provider struct {
	Client             *auth.Client
	oauthStateString   string
	WebApiKey          string
	RedirectURL        string
	SendOobCodeURL     string
	SignInWithEmailURL string
	App                *firebase.App
}

func New(webapikey string, credentials string, redirect string) Provider {
	fmt.Println(credentials, webapikey)
	opt := option.WithCredentialsFile(credentials)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error initializing auth client: %v\n", err)
	}
	fmt.Println("vse ok")
	return Provider{
		oauthStateString:   randfuncs.RandomString(20, randfuncs.NewSource()),
		Client:             client,
		WebApiKey:          webapikey,
		RedirectURL:        redirect,
		SendOobCodeURL:     "https://identitytoolkit.googleapis.com/v1/accounts:sendOobCode",
		SignInWithEmailURL: "https://identitytoolkit.googleapis.com/v1/accounts:signInWithEmailLink",
		App:                app,
	}
}
func (p Provider) BeginAuth(w http.ResponseWriter, r *http.Request) {
	form := url.Values{}
	form.Add("requestType", "EMAIL_SIGNIN")
	form.Add("email", r.FormValue("email"))
	form.Add("continueUrl", p.RedirectURL)
	req, err := http.NewRequest("POST", p.SendOobCodeURL+"?key="+p.WebApiKey, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log.Println(err, "creating request error")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err, "request error")
	}

	c := &http.Cookie{
		Name:     "email",
		Value:    r.FormValue("email"),
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, c)

	err = resp.Write(w)
	if err != nil {
		log.Println(err, "write error")
	}

}

type firebaseResponse struct {
	IdToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
}

func (p Provider) CompleteAuth(w http.ResponseWriter, r *http.Request) {
	tokens := &firebaseResponse{}
	c, err := r.Cookie("email")
	if err != nil {
		log.Println(err, "cookie retrieve error")
	}

	oobCode := r.URL.Query().Get("oobCode")

	form := url.Values{}
	form.Add("oobCode", oobCode)
	form.Add("email", c.Value)
	req, err := http.NewRequest("POST", p.SignInWithEmailURL+"?key="+p.WebApiKey, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err, "request error")
	}
	err = json.NewDecoder(resp.Body).Decode(tokens)
	if err != nil {
		log.Println(err, "json decode error")
	}

	idToken := http.Cookie{Name: "idToken", Value: tokens.IdToken, MaxAge: 3600, Path: "/", HttpOnly: true, Secure: true}
	refreshToken := http.Cookie{Name: "refreshToken", Value: tokens.RefreshToken, Path: "/refresh", HttpOnly: true, Secure: true}
	http.SetCookie(w, &idToken)
	http.SetCookie(w, &refreshToken)
	c = &http.Cookie{
		Name:     "email",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	}
	http.SetCookie(w, c)

	err = resp.Write(w)
	if err != nil {
		log.Println(err, "write error")
	}

}
