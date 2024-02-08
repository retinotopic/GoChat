package auth

import (
	"net/http"

	"github.com/gorilla/mux"
)

type LoginCreator interface {
	BeginLoginCreator
	CompleteLoginCreator
}
type BeginLoginCreator interface {
	BeginLoginCreate(w http.ResponseWriter, r *http.Request)
}
type CompleteLoginCreator interface {
	CompleteLoginCreate(w http.ResponseWriter, r *http.Request)
}

func BeginAuthRoute(w http.ResponseWriter, r *http.Request) {
	provider := mux.Vars(r)["provider"]
	if provider == "" {
		http.Error(w, "Provider not specified", http.StatusBadRequest)
		return
	}
	if _, ok := Providersmap[provider]; !ok {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}
	Providersmap[provider].BeginLoginCreate(w, r)
}
func CompleteAuthRoute(w http.ResponseWriter, r *http.Request) {
	provider := mux.Vars(r)["provider"]
	if provider == "" {
		http.Error(w, "Provider not specified", http.StatusBadRequest)
	}
	if _, ok := Providersmap[provider]; !ok {
		http.Error(w, "Provider not found", http.StatusNotFound)
	}
	Providersmap[provider].CompleteLoginCreate(w, r)
}
