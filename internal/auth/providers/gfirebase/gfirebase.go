package gfirebase

import (
	"context"
	"log"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/retinotopic/pokerGO/pkg/randfuncs"
	"google.golang.org/api/option"
)

type Provider struct {
	Client           *auth.Client
	oauthStateString string
	WebApiKey        string
	Settings         *auth.ActionCodeSettings
}

func New(webapikey string, credentials string, redirect string) Provider {
	opt := option.WithCredentialsJSON([]byte(credentials))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	return Provider{
		oauthStateString: randfuncs.RandomString(20, randfuncs.NewSource()),
		Client:           client,
		WebApiKey:        webapikey,
		Settings:         &auth.ActionCodeSettings{URL: redirect},
	}
}
func (p Provider) BeginAuth(w http.ResponseWriter, r *http.Request) {
	url, err := p.Client.EmailSignInLink(context.Background(), r.FormValue("email"), p.Settings)
	if err != nil {
		log.Println(err, "url compose firebase")
	}
	w.Write([]byte(url))
}
func (p Provider) CompleteAuth(w http.ResponseWriter, r *http.Request) {

}
