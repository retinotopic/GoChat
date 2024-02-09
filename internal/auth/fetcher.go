package auth

import (
	"log"
	"net/http"
)

type Fetcher interface {
	FetchUser(w http.ResponseWriter, r *http.Request)
}

func FetchUserMW(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("provider")
	if err != nil {
		log.Println(err, "revoke cookie retrieve err")
	}

	Providersmap[c.Value].FetchUser(w, r)
}
