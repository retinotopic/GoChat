package auth

import (
	"net/http"
)

type BeginLoginCreator interface {
	BeginLoginCreate(w http.ResponseWriter, r *http.Request)
}
type CompleteLoginCreator interface {
	CompleteLoginCreate(w http.ResponseWriter, r *http.Request)
}

func BeginLoginCreateI(beginLoginCreator BeginLoginCreator) http.Handler {
	return http.HandlerFunc(beginLoginCreator.BeginLoginCreate)
}
func CompleteLoginCreateI(completeLoginCreator CompleteLoginCreator) http.Handler {
	return http.HandlerFunc(completeLoginCreator.CompleteLoginCreate)
}

/*func CompleteLoginCreateI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		completeLoginCreator := r.Context().Value("completeLoginCreator").(CompleteLoginCreator)
		completeLoginCreator.CompleteLoginCreate(w, r)
	})
}*/
