package google

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pascaldekloe/jwt"
	"golang.org/x/oauth2/google"
)

func (p Provider) Refresh(w http.ResponseWriter, r *http.Request) {
	form := url.Values{}
	token, err := r.Cookie("refreshToken")
	if err != nil {
		log.Println(err, "revoke cookie retrieve err")
	}
	form.Add("refresh_token", token.Value)
	form.Add("grant_type", "refresh_token")
	form.Add("client_id", p.Config.ClientID)
	form.Add("client_secret", p.Config.ClientSecret)
	req, err := http.NewRequest("POST", google.Endpoint.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		log.Println(err, "error creating request error")
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err, "error request error")
	}
	log.Println(resp.StatusCode)

}
func (p Provider) Revoke(w http.ResponseWriter, r *http.Request) {
	form := url.Values{}
	token, err := r.Cookie("token")
	if err != nil {
		log.Println(err, "revoke cookie retrieve err")
	}
	form.Add("token", token.Value)
	req, err := http.NewRequest("POST", p.RevokeURL, strings.NewReader(form.Encode()))
	if err != nil {
		log.Println(err, "error request error")
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err, "error request error")
	}
	log.Println(resp.StatusCode)

}
func (p Provider) FetchUser(w http.ResponseWriter, r *http.Request) {
	block, err := pem.Decode([]byte(p.PublicKey))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
	}
	var cert *x509.Certificate
	cert, err2 := x509.ParseCertificate(block.Bytes)
	rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)
	if err2 != nil {
		w.WriteHeader(http.StatusUnauthorized)
	}
	token, err3 := r.Cookie("token")
	if err3 != nil {
		w.WriteHeader(http.StatusUnauthorized)
	}
	claims, err4 := jwt.RSACheck([]byte(token.Value), rsaPublicKey)
	if err4 != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	if !claims.Valid(time.Now()) {
		w.WriteHeader(http.StatusBadRequest)
	}
}
