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
	PublicKey        string
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
		PublicKey:        "-----BEGIN CERTIFICATE-----\nMIIDJjCCAg6gAwIBAgIIbUDYzBFHdOQwDQYJKoZIhvcNAQEFBQAwNjE0MDIGA1UE\nAwwrZmVkZXJhdGVkLXNpZ25vbi5zeXN0ZW0uZ3NlcnZpY2VhY2NvdW50LmNvbTAe\nFw0yNDAxMzEwNDM4MTVaFw0yNDAyMTYxNjUzMTVaMDYxNDAyBgNVBAMMK2ZlZGVy\nYXRlZC1zaWdub24uc3lzdGVtLmdzZXJ2aWNlYWNjb3VudC5jb20wggEiMA0GCSqG\nSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC/d1kvZHY+55tsAKFhcpVbpH4UkNRWZnxP\nJUxYfT5WlWMVhm/LsFloTkCktZzdSGle6DqvrbfQAkjSjDnJ4+eqZmcjmpyTHPRu\nvQWPbtl2D5fg9Y33mB2Tp+kjgnA2ZkWyCJbOYOI/XXyPyMjEw0GIhU9PtHLKSEBl\n42cYLaQNT7zWKinznQrbTkTB9KL9MFPocJtGP9NJDagl8hcM9fyzsqDg9GMM4e3c\nPwKKqwhZvFKRFG5OJT/UCGxu5zd32GQPWs45OFVPpPstVlPvXRa09rVBspSAYi7a\nmaI+jQEJ2duqjOxFU7Bj3TVHvtWVXClO4aic9m7I7UHYWAa5iCi3AgMBAAGjODA2\nMAwGA1UdEwEB/wQCMAAwDgYDVR0PAQH/BAQDAgeAMBYGA1UdJQEB/wQMMAoGCCsG\nAQUFBwMCMA0GCSqGSIb3DQEBBQUAA4IBAQAU1vWzq3YYDQ5P1adxjfs9MYv4s/zN\npS8mQqGN0w27X9wcy5Jix+2Of85fvt4eSkBZgbGeRIfr+Omy8qr9zefwpFoQPkvh\nwM2JDRucjfTPjlhkLFt8yr0ZwwBGziWpCVFBjZhpHdDynhfNaI+RFlO2XHnXuQRS\n4B9c9JyaaaOkPu+XIgb5zP5AIUepVgpDdF/nPWeSBbIBoZLZR5XFCmRW55tVtubp\nBaui/Yclnr/36NTKK8IzOLE/ha5PuQH2Ai3WxbztIresBdadjhKxm7iOz63akWPy\nSJ5/l4Zn2MEu6JLTaVriZMmOmhNLhfmJFO5dTZpA4Vfd5xaN5VG39kjW\n-----END CERTIFICATE-----\n",
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
	token, err := p.Config.Exchange(context.Background(), code)
	if err != nil {
		fmt.Println(err, "exchange error")
	}
	jwt_token := http.Cookie{Name: "jwt_token", Value: token.Extra("id_token").(string), MaxAge: 3600, Path: "/", HttpOnly: true, Secure: true}
	refresh_token := http.Cookie{Name: "refresh_token", Value: token.RefreshToken, Path: "/refresh", HttpOnly: true, Secure: true}
	http.SetCookie(w, &refresh_token)
	http.SetCookie(w, &jwt_token)

	//////////////
	fmt.Println(token.Expiry)
	fmt.Println(token.Extra("id_token"), "extra")
	r.Header.Add("Authorization", "Bearer "+token.Extra("id_token").(string))
	resp2, _ := http.DefaultClient.Do(r)
	_ = resp2.Write(w)
}
