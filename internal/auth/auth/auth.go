package auth

import (
	"net/http"
	"os"

	"github.com/retinotopic/GoChat/internal/auth/providers/google"
	"github.com/retinotopic/GoChat/internal/auth/providers/stytch"
)

type Providers map[string]Authenticator

var Providersmap = Providers{
	"google": google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), "http://localhost:8080/google/CompleteAuth"),
	"stytch": stytch.New(os.Getenv("STYTCH_PROJECT_ID"), os.Getenv("STYTCH_PROJECT_SECRET"), "http://localhost:8080/stytch/CompleteAuth"),
}

type Authenticator interface {
	BeginAuth(w http.ResponseWriter, r *http.Request)
	CompleteAuth(w http.ResponseWriter, r *http.Request)
}
