package auth

import (
	"net/http"
)

type Revoker interface {
	Revoke(w http.ResponseWriter, r *http.Request)
}
type Refresher interface {
	Refresh(w http.ResponseWriter, r *http.Request)
}

func RefreshI(refresher Refresher) http.Handler {
	return http.HandlerFunc(refresher.Refresh)
}

func RevokeI(revoker Revoker) http.Handler {
	return http.HandlerFunc(revoker.Revoke)
}
