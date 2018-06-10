package gohttpmw

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type augmentedResponseWriter struct {
	http.ResponseWriter
	length     int
	httpStatus int
}

// WriteHeader will not only write b to w
// but also save the http status in the struct
func (w *augmentedResponseWriter) WriteHeader(httpStatus int) {
	w.ResponseWriter.WriteHeader(httpStatus)
	w.httpStatus = httpStatus
}

// Write will not only write b to w but also save the byte length in the struct
func (w *augmentedResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length = n

	return n, err
}

func newAugmentedResponseWriter(
	w http.ResponseWriter,
) *augmentedResponseWriter {
	return &augmentedResponseWriter{
		ResponseWriter: w,
		httpStatus:     http.StatusOK,
	}
}

// Logger will return an error if the required params are not there
func Logger(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			startTime := time.Now()
			naw := newAugmentedResponseWriter(w)
			h.ServeHTTP(naw, r)

			if reqID := GetRequestID(r.Context()); reqID != "" {
				logger = logger.With("request_id", reqID)
			}

			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			logger = logger.With(
				"http_scheme", scheme,
				"http_proto", r.Proto,
				"http_method", r.Method,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"host", r.Host,
				"uri", r.RequestURI,
				"process_time", time.Since(startTime),
				"http_status", naw.httpStatus,
				"resp_length", naw.length,
			)

			reqErr := GetRequestError(r.Context())
			if reqErr != nil {
				// Get response status and size
				if naw.httpStatus == http.StatusInternalServerError {
					logger.Error(reqErr)
					return
				}
				logger.Warn(reqErr)
				return
			}

			logger.Info()

		})
	}
}

// ContextKeyRequestError will allow the error to be passed down
const ContextKeyRequestError = ContextKey("requestError")

// GetRequestError will retrieve the request error
// from the context if there is one
func GetRequestError(ctx context.Context) error {
	if reqErr, ok := ctx.Value(ContextKeyRequestError).(error); ok {
		return reqErr
	}

	return nil
}

// SetRequestError sets the error in the context so it can be picked up
// for logging
func SetRequestError(r *http.Request, err error) {
	*r = *r.WithContext(
		context.WithValue(r.Context(), ContextKeyRequestError, err),
	)
}
