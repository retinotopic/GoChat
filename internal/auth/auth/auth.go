package auth

import (
	"os"

	"github.com/retinotopic/pokerGO/pkg/randfuncs"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	GoogleConfig     *AuthConfig
	DiscordConfig    *AuthConfig
	OauthStateString string
)

type info struct {
	Access_token  string `json:"access_token"`
	Refresh_token string `json:"refresh_token"`
	Id            string `json:"id"`
}
type AuthConfig struct {
	Config      *oauth2.Config
	revokeURL   string
	userinfoURL string
}

func init() {
	OauthStateString = randfuncs.RandomString(20, randfuncs.NewSource())
	GoogleConfig = &AuthConfig{
		Config: &oauth2.Config{
			RedirectURL:  "http://localhost:8080/callback",
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			Scopes:       []string{"email"},
			Endpoint:     google.Endpoint,
		},
		revokeURL:   "https://accounts.google.com/o/oauth2/revoke",
		userinfoURL: "https://www.googleapis.com/oauth2/v3/userinfo",
	}
	DiscordConfig = &AuthConfig{
		Config: &oauth2.Config{
			RedirectURL:  "http://localhost:8080/callback",
			ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
			ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
			Scopes:       []string{"identify"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://discord.com/oauth2/authorize",
				TokenURL: "https://discord.com/api/oauth2/token",
			},
		},
		revokeURL:   "https://discord.com/api/oauth2/token/revoke",
		userinfoURL: "https://discord.com/api/users/@me",
	}
}
