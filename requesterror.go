package gohttpmw

import (
	"context"
	"net/http"
)

// ContextKeyRequestError will allow the error to be passed down
const ContextKeyRequestError = ContextKey("requestError")

// SetRequestError sets the error in the context so it can be picked up
// for logging
func SetRequestError(r *http.Request, err error) {
	*r = *r.WithContext(
		context.WithValue(r.Context(), ContextKeyRequestError, err),
	)
}

// GetRequestError will retrieve the request error
// from the context if there is one
func GetRequestError(ctx context.Context) error {
	if reqErr, ok := ctx.Value(ContextKeyRequestError).(error); ok {
		return reqErr
	}

	return nil
}
