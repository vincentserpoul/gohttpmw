package gohttpmw

import (
	"context"
	"net/http"

	"github.com/rs/xid"
)

const (
	// ContextKeyRequestID allow storage of requestid in the context
	ContextKeyRequestID = ContextKey("requestID")
)

// RequestID adds a requestID to the request context
// ksuid is a unique global id that is orderable by time (a step up normal uuid)
func RequestID() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := xid.New()
			w.Header().Add("requestID", requestID.String())
			ctx := context.WithValue(
				r.Context(),
				ContextKeyRequestID,
				requestID,
			)
			// We don't want to lose the reference to the Request
			*r = *r.WithContext(ctx)
			h.ServeHTTP(w, r)
		})
	}
}

// GetRequestID will retrieve the request id from the context if there is one
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return reqID
	}

	return ""
}
