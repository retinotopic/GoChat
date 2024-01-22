package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/retinotopic/pokerGO/pkg/randfuncs"
	"golang.org/x/oauth2"
)

var (
	OauthStateString string
	ProvidersConfig  providersConfig
)

type info struct {
	Access_token  string `json:"access_token"`
	Refresh_token string `json:"refresh_token"`
	Id            string `json:"id"`
}
type providersConfig map[string]authConfig

type authConfig struct {
	Config      *oauth2.Config
	RevokeURL   string
	UserinfoURL string
}
type providerJSON struct {
	Name         string `json:"Name"`
	RedirectURL  string `json:"RedirectURL"`
	ClientID     string `json:"ClientID"`
	ClientSecret string `json:"ClientSecret"`
	Scopes       string `json:"Scopes"`
	AuthURL      string `json:"AuthURL"`
	TokenURL     string `json:"TokenURL"`
	RevokeURL    string `json:"RevokeURL"`
	UserinfoURL  string `json:"UserinfoURL"`
}

func init() {
	OauthStateString = randfuncs.RandomString(20, randfuncs.NewSource())
	jsonFile, err := os.Open("config.json")
	jsonByte, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err, "readall err")
	}
	var providersJSON map[string]providerJSON
	err = json.Unmarshal(jsonByte, &providersJSON)
	if err != nil {
		fmt.Println(err, "unM json")
	}
	ProvidersConfig = make(providersConfig)
	for _, v := range providersJSON {
		ProvidersConfig[v.Name] = authConfig{
			Config: &oauth2.Config{
				ClientID:     v.ClientID,
				ClientSecret: v.ClientSecret,
				RedirectURL:  v.RedirectURL,
				Scopes:       []string{v.Scopes},
				Endpoint: oauth2.Endpoint{
					AuthURL:  v.AuthURL,
					TokenURL: v.TokenURL,
				},
			},
			RevokeURL:   v.RevokeURL,
			UserinfoURL: v.UserinfoURL,
		}

	}
	OauthStateString = randfuncs.RandomString(20, randfuncs.NewSource())

}
