package middleware

import (
	"net/http"
)

type UserMiddleware struct {
	FetchUser func(http.ResponseWriter, *http.Request) (string, error)
}

func (um *UserMiddleware) GetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sub, err := um.FetchUser(w, r)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

		}

		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), sub)))
	})
}
