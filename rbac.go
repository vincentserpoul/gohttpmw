package gohttpmw

import (
	"context"
	"net/http"

	"github.com/ory/ladon"
)

// RBAC checks if the user is allowed to do the request
func RBAC(
	warden ladon.Warden,
	getRoleFunc func(context.Context) string,
) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := warden.IsAllowed(&ladon.Request{
				Subject:  getRoleFunc(r.Context()),
				Action:   r.Method,
				Resource: r.RequestURI,
			}); err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
