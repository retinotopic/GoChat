package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/retinotopic/tinyauth/provider"
)

type ProviderMap map[string]provider.Provider

func (pm ProviderMap) getProvider(r *http.Request) (provider.Provider, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return nil, err
	}
	token, _, err := new(jwt.Parser).ParseUnverified(cookie.Value, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	var issuer string
	var ok bool
	var claims jwt.MapClaims
	var iss interface{}
	if claims, ok = token.Claims.(jwt.MapClaims); !ok {
		return nil, fmt.Errorf("issuer claim retrieve error")
	}
	if iss, ok = claims["iss"]; !ok {
		return nil, fmt.Errorf("issuer claim retrieve error")
	}
	if issuer, ok = iss.(string); !ok {
		return nil, fmt.Errorf("issuer claim is not a string")
	}

	if prvdr, ok := pm[issuer]; ok {
		return prvdr, nil
	}
	return nil, errors.New("no such provider")
}
func (pm ProviderMap) BeginAuth(w http.ResponseWriter, r *http.Request) {
	prvdr, err := pm.getProvider(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	prvdr.BeginAuth(w, r)
}
func (pm ProviderMap) CompleteAuth(w http.ResponseWriter, r *http.Request) {
	prvdr, err := pm.getProvider(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	prvdr.CompleteAuth(w, r)
}
func (pm ProviderMap) FetchUser(w http.ResponseWriter, r *http.Request) (string, error) {
	prvdr, err := pm.getProvider(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	return prvdr.FetchUser(w, r)
}
