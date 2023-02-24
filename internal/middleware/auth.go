package middleware

import (
	"fmt"
	"gomarket/internal/cookies"
	"net/http"
)

func AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := cookies.Get(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		}

		if !cookies.Check(cookie) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Sprint(`{"error": "bad cookie"}`)))
			return
		}

		next.ServeHTTP(w, r)
	})
}
