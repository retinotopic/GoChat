package middleware

import (
	"net/http"
)

type Fetcher interface {
	FetchUser(http.ResponseWriter, *http.Request) (string, error)
}
type UserMiddleware struct {
	Fetcher
}

func (u *UserMiddleware) GetUserMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sub, err := u.FetchUser(w, r)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), sub)))
	})
}
