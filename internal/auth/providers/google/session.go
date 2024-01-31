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
)

func (p Provider) Refresh(w http.ResponseWriter, r *http.Request) {
	form := url.Values{}
	token, err := r.Cookie("refresh_token")
	if err != nil {
		log.Println(err, "revoke cookie retrieve err")
	}
	form.Add("refresh_token", token.Value)
	form.Add("grant_type", "refresh_token")
	form.Add("client_id", p.Config.ClientID)
	form.Add("client_secret", p.Config.ClientSecret)
	req, _ := http.NewRequest("POST", "https://login.microsoftonline.com/TenantID/v2.0/oauth2/token", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp2, _ := http.DefaultClient.Do(req)
	_ = resp2.Write(w)

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
	resp2, _ := http.DefaultClient.Do(req)
	err = resp2.Write(w)
	if err != nil {
		log.Println(err, "error response error")
	}
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
