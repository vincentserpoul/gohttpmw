package gohttpmw

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// LoggerZero will log the full request details, with performance
func LoggerZero(logger zerolog.Logger) func(http.Handler) http.Handler {
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
