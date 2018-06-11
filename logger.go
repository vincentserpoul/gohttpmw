package gohttpmw

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
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
func Logger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			startTime := time.Now()
			naw := newAugmentedResponseWriter(w)
			h.ServeHTTP(naw, r)

			// Copy the context into a new logger
			l := logger.With().Logger()

			if reqID := GetRequestID(r.Context()); reqID != "" {
				l = l.With().Str("request_id", reqID).Logger()
			}

			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			l = l.With().
				Str("http_scheme", scheme).
				Str("http_proto", r.Proto).
				Str("http_method", r.Method).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Str("host", r.Host).
				Str("uri", r.RequestURI).
				Dur("process_time", time.Since(startTime)).
				Int("http_status", naw.httpStatus).
				Int("resp_length", naw.length).Logger()

			reqErr := GetRequestError(r.Context())
			if reqErr != nil {
				// Get response status and size
				if naw.httpStatus == http.StatusInternalServerError {
					l.Error().Msg(reqErr.Error())
					return
				}
				l.Warn().Msg(reqErr.Error())
				return
			}

			l.Info().Msg("")
		})
	}
}
