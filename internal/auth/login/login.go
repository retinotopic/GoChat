package login

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/retinotopic/GoChat/internal/auth/auth"
)

func LoginUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("im here1")
	url := auth.ProvidersConfig[r.Header.Get("Provider")].Config.AuthCodeURL(auth.OauthStateString, nil)
	fmt.Println(url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect) // composing our auth request url
}
func Callback(w http.ResponseWriter, r *http.Request) {
	content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"), r.Header.Get("Provider"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	fmt.Fprintf(w, "Content: %s\n", content)
}
func getUserInfo(state string, code string, provider string) ([]byte, error) {
	if state != auth.OauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := auth.ProvidersConfig[provider].Config.Exchange(context.Background(), code)
	fmt.Println(token.RefreshToken)
	fmt.Println(token.AccessToken)
	fmt.Println(token.Expiry)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}
	return contents, nil
}
