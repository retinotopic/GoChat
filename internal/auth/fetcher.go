package auth

import (
	"net/http"
)

type Fetcher interface {
	FetchUser(w http.ResponseWriter, r *http.Request)
}

func FetchUserMW(fetcher Fetcher) http.Handler {
	return http.HandlerFunc(fetcher.FetchUser)
}
