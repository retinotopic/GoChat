package auth

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Authenticator interface {
	BeginAuth(w http.ResponseWriter, r *http.Request)
	CompleteAuth(w http.ResponseWriter, r *http.Request)
}

func BeginAuthenticator(w http.ResponseWriter, r *http.Request) {
	provider := mux.Vars(r)["provider"]
	if provider == "" {
		http.Error(w, "Provider not specified", http.StatusBadRequest)
		return
	}
	if _, ok := Providersmap[provider]; !ok {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}
	Providersmap[provider].BeginAuth(w, r)
}
func CompleteAuthenticator(w http.ResponseWriter, r *http.Request) {
	provider := mux.Vars(r)["provider"]
	if provider == "" {
		http.Error(w, "Provider not specified", http.StatusBadRequest)
	}
	if _, ok := Providersmap[provider]; !ok {
		http.Error(w, "Provider not found", http.StatusNotFound)
	}
	Providersmap[provider].CompleteAuth(w, r)
}
