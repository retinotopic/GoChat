package auth

import (
	"net/http"

	"github.com/gorilla/mux"
)

type RevokeRefresher interface {
	Revoke(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
}

func RefreshRoute(w http.ResponseWriter, r *http.Request) {
	provider := mux.Vars(r)["provider"]
	if provider == "" {
		http.Error(w, "Provider not specified", http.StatusBadRequest)
		return
	}
	if _, ok := Providersmap[provider]; !ok {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}
	Providersmap[provider].Refresh(w, r)
}
func RevokeRoute(w http.ResponseWriter, r *http.Request) {
	provider := mux.Vars(r)["provider"]
	if provider == "" {
		http.Error(w, "Provider not specified", http.StatusBadRequest)
	}
	if _, ok := Providersmap[provider]; !ok {
		http.Error(w, "Provider not found", http.StatusNotFound)
	}
	Providersmap[provider].Revoke(w, r)
}
