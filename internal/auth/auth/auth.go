package auth

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/retinotopic/GoChat/internal/auth/providers/google"
)

type Providers map[string]Authenticator

var providersmap = Providers{
	"google": google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), "http://localhost:8080/google/CompleteAuth"),
} //"gfirebase":os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

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
	if _, ok := providersmap[provider]; !ok {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}
	providersmap[provider].BeginAuth(w, r)
}
func CompleteAuthenticator(w http.ResponseWriter, r *http.Request) {
	provider := mux.Vars(r)["provider"]
	if provider == "" {
		http.Error(w, "Provider not specified", http.StatusBadRequest)
	}
	if _, ok := providersmap[provider]; !ok {
		http.Error(w, "Provider not found", http.StatusNotFound)
	}
	providersmap[provider].CompleteAuth(w, r)
}
