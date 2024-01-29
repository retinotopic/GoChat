package stytch

import (
	"context"
	"log"
	"net/http"

	"github.com/retinotopic/pokerGO/pkg/randfuncs"
	"github.com/stytchauth/stytch-go/v12/stytch/consumer/otp"
	"github.com/stytchauth/stytch-go/v12/stytch/consumer/otp/email"
	"github.com/stytchauth/stytch-go/v12/stytch/consumer/stytchapi"
)

type Provider struct {
	Config           *stytchapi.API
	oauthStateString string
}

func New(projectid string, projectsecret, redirect string) Provider {
	client, err := stytchapi.NewClient(
		projectid,
		projectsecret,
	)
	if err != nil {
		log.Fatalf("error instantiating API client %s", err)
	}
	return Provider{
		Config:           client,
		oauthStateString: randfuncs.RandomString(20, randfuncs.NewSource()),
	}
}
func (p Provider) BeginAuth(w http.ResponseWriter, r *http.Request) {
	params := &email.LoginOrCreateParams{
		Email: r.FormValue("email"),
	}
	resp, err := p.Config.OTPs.Email.LoginOrCreate(context.Background(), params)
	if err != nil {
		log.Println(err)
	}
	log.Println(resp)
}
func (p Provider) CompleteAuth(w http.ResponseWriter, r *http.Request) {
	params := &otp.AuthenticateParams{
		MethodID: "phone-number-test-d5a3b680-e8a3-40c0-b815-ab79986666d0",
		Code:     r.FormValue("code"),
	}

	resp, err := p.Config.OTPs.Authenticate(context.Background(), params)
	if err != nil {
		log.Println(err)
	}

	log.Println(resp)
}
