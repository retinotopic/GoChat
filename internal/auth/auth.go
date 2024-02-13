package auth

type Providers map[string]LoginFetchRevokeRefresher

/*
	var providers = Providers{
		"google":    google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), "http://localhost:8080/google/CompleteAuth"),
		"gfirebase": gfirebase.New(os.Getenv("WEB_API_KEY"), os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), "http://localhost:8080/gfirebase/CompleteAuth"),
	}
*/
var CurrentProviders = Providers{}

type LoginFetchRevokeRefresher interface {
	LoginCreator
	RevokeRefresher
	Fetcher
}
type LoginCreator interface {
	BeginLoginCreator
	CompleteLoginCreator
}
type RevokeRefresher interface {
	Revoker
	Refresher
}
