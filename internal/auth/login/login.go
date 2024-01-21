package login

import (
	"fmt"
	"net/http"

	"github.com/retinotopic/GoChat/internal/auth/auth"
)

func LoginUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("im here1")
	url := auth.GoogleConfig.Config.AuthCodeURL(auth.OauthStateString, nil)
	fmt.Println(url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect) // composing our auth request url
}
