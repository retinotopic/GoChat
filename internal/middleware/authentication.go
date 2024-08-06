package middleware

import (
	"net/http"

	"github.com/retinotopic/tinyauth/provider"
)

type UserMiddleware struct {
	GetUser     func(string) (int, string, error)
	GetProvider func(http.ResponseWriter, *http.Request) (provider.Provider, error)
}

func (um *UserMiddleware) FetchUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prvdr, err := um.GetProvider(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		sub, err := prvdr.FetchUser(w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

		}

		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), sub)))
	})
}
