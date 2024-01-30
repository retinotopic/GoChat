package google

import (
	"context"
	"fmt"
	"net/http"

	"github.com/retinotopic/pokerGO/pkg/randfuncs"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Provider struct {
	Name             string
	Config           oauth2.Config
	RevokeURL        string
	oauthStateString string
}

func New(clientid string, clientsecret, redirect string) Provider {
	return Provider{
		Config: oauth2.Config{
			ClientID:     clientid,
			ClientSecret: clientsecret,
			RedirectURL:  redirect,
			Scopes:       []string{"email openid"},
			Endpoint:     google.Endpoint,
		},
		RevokeURL:        "https://accounts.google.com/o/oauth2/revoke",
		oauthStateString: randfuncs.RandomString(20, randfuncs.NewSource()),
	}
}

func (p Provider) BeginAuth(w http.ResponseWriter, r *http.Request) {
	url := p.Config.AuthCodeURL(p.oauthStateString, oauth2.AccessTypeOffline)
	fmt.Println(url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect) // composing our auth request url
}

func (p Provider) CompleteAuth(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != p.oauthStateString {
		fmt.Println("invalid oauth state")
	}

	code := r.FormValue("code")
	token, _ := p.Config.Exchange(context.Background(), code)
	//client := config.Client(context.Background(), token)
	//resp,_:= client.Get("https://discord.com/api/users/@me")
	fmt.Println(token.AccessToken)
	fmt.Println(token.RefreshToken)
	fmt.Println(token.Expiry)
	fmt.Println(token.Extra("id_token"), "extra")
	r.Header.Add("Authorization", "Bearer "+token.Extra("id_token").(string))
	resp2, _ := http.DefaultClient.Do(r)
	_ = resp2.Write(w)
}
