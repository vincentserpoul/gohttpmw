package gohttpmw

import (
	"net/http"
)

// CSP adds a CSP header to the response writer
func CSP(c string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("content-security-policy", c)
			// We don't want to lose the reference to the Request
			h.ServeHTTP(w, r)
		})
	}
}
