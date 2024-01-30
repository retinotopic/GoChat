package google

import (
	"log"
	"net/http"
	"net/url"
	"strings"
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
