package gohttpmw

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger will run the full request details
// if you need performance, look into loggerZero
func Logger(l *logrus.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			naw := newAugmentedResponseWriter(w)
			startTime := time.Now()

			h.ServeHTTP(naw, r)

			logFields := logrus.Fields{}

			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			logFields["http_scheme"] = scheme
			logFields["http_proto"] = r.Proto
			logFields["http_method"] = r.Method

			logFields["remote_addr"] = r.RemoteAddr
			logFields["user_agent"] = r.UserAgent()

			logFields["host"] = r.Host
			logFields["uri"] = r.RequestURI

			logFields["process_time"] = float64(
				time.Since(startTime) / time.Millisecond,
			)
			logFields["http_status"] = naw.httpStatus
			logFields["resp_length"] = naw.length

			if reqID := GetRequestID(r.Context()); reqID != "" {
				logFields["request_id"] = reqID
			}

			// Get additional logging fields
			atl := GetAddToRequestLog(r.Context())
			for k, v := range atl {
				logFields[k] = v
			}

			reqErr := GetRequestError(r.Context())
			if reqErr != nil {
				// Get response status and size
				if naw.httpStatus == http.StatusInternalServerError {
					l.WithFields(logFields).Errorln(reqErr.Error())
					return
				}
				l.WithFields(logFields).Warnln(reqErr.Error())
				return
			}

			l.WithFields(logFields).Infoln()
		})
	}
}

const (
	// ContextKeyAddToRequestLog allow storage of additional log fields in the context
	ContextKeyAddToRequestLog = ContextKey("AddToLog")
)

// AddToRequestLog allows to add more fields to the request log
func AddToRequestLog(
	k string,
	f func(context.Context) interface{},
) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a := GetAddToRequestLog(r.Context())
			if v := f(r.Context()); v != nil {
				a[k] = v
			}
			ctx := context.WithValue(
				r.Context(),
				ContextKeyAddToRequestLog,
				a,
			)
			*r = *r.WithContext(ctx)
			h.ServeHTTP(w, r)
		})
	}
}

// GetAddToRequestLog will retrieve the fileds to be added to the log
func GetAddToRequestLog(ctx context.Context) map[string]interface{} {
	if addToLog, ok := ctx.Value(
		ContextKeyAddToRequestLog,
	).(map[string]interface{}); ok {
		return addToLog
	}

	return make(map[string]interface{})
}
