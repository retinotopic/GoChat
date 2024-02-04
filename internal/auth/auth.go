package auth

import (
	"os"

	"github.com/retinotopic/GoChat/internal/providers/gfirebase"
	"github.com/retinotopic/GoChat/internal/providers/google"
)

type Providers map[string]AuthFetchRevokeRefresher

var Providersmap = Providers{
	"google":    google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), "http://localhost:8080/google/CompleteAuth"),
	"gfirebase": gfirebase.New(os.Getenv("WEB_API_KEY"), os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), "http://localhost:8080/gfirebase/CompleteAuth"),
}

type AuthFetchRevokeRefresher interface {
	Authenticator
}
