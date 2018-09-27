package gohttpmw

import (
	"net/http"
)

// Security adds secure headers
func Security() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Frame-Options", "SAMEORIGIN")
			w.Header().Add("X-Content-Type-Options", "nosniff")
			w.Header().Add("X-XSS-Protection", "1; mode=block")
			w.Header().Add("Referrer-Policy", "same-origin")

			h.ServeHTTP(w, r)
		})
	}
}
