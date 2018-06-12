package gohttpmw

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

// nolint[:gocyclo]
func TestLoggerZero(t *testing.T) {
	errTest := errors.New("test error")

	tc := []struct {
		name               string
		handler            http.HandlerFunc
		expectedLogLevel   zerolog.Level
		expectedHTTPStatus int
		withHTTPS          bool
		withRequestID      bool
	}{
		{
			name: "classic request log",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			expectedLogLevel:   zerolog.InfoLevel,
			expectedHTTPStatus: http.StatusOK,
		},
		{
			name: "request log with error",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					*req = *req.WithContext(
						context.WithValue(
							req.Context(),
							ContextKeyRequestError,
							errTest,
						),
					)
					w.WriteHeader(http.StatusInternalServerError)
				}),
			expectedLogLevel:   zerolog.ErrorLevel,
			expectedHTTPStatus: http.StatusInternalServerError,
		},
		{
			name: "request log with error as warning",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					*req = *req.WithContext(
						context.WithValue(
							req.Context(),
							ContextKeyRequestError,
							errTest,
						),
					)
					w.WriteHeader(http.StatusBadRequest)
				}),
			expectedLogLevel:   zerolog.WarnLevel,
			expectedHTTPStatus: http.StatusBadRequest,
		},
		{
			name: "request log with requestID",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					*req = *req.WithContext(
						context.WithValue(
							req.Context(),
							ContextKeyRequestID,
							xid.New().String(),
						),
					)
					w.WriteHeader(http.StatusOK)
				}),
			expectedLogLevel:   zerolog.InfoLevel,
			expectedHTTPStatus: http.StatusOK,
			withRequestID:      true,
		},
		{
			name: "404 request log",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}),
			expectedLogLevel:   zerolog.InfoLevel,
			expectedHTTPStatus: http.StatusNotFound,
		},
		{
			name: "400 request log",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				}),
			expectedLogLevel:   zerolog.InfoLevel,
			expectedHTTPStatus: http.StatusBadRequest,
		},
		{
			name: "https request log",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			expectedLogLevel:   zerolog.InfoLevel,
			withHTTPS:          true,
			expectedHTTPStatus: http.StatusOK,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			midWared := LoggerZero(zerolog.New(out))(tt.handler)
			rr := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = "127.0.0.1"
			r.Header.Set("User-Agent", "test")
			if tt.withHTTPS {
				r.TLS = &tls.ConnectionState{}
			}
			midWared.ServeHTTP(rr, r)

			logRes := make(map[string]interface{})
			if err := json.Unmarshal(out.Bytes(), &logRes); err != nil {
				t.Fatalf("error unmarshalling log %v", err)
				return
			}

			if tt.expectedLogLevel.String() != logRes["level"].(string) {
				t.Errorf(
					"wrong log level, expected %s, got %s",
					tt.expectedLogLevel.String(), logRes["level"].(string),
				)
				return
			}

			// Check all existing fields
			if float64(tt.expectedHTTPStatus) != logRes["http_status"].(float64) {
				t.Errorf(
					"wrong httpstatus, expected %d, got %f",
					tt.expectedLogLevel,
					logRes["http_status"].(float64),
				)
				return
			}

			if tt.withRequestID {
				val, ok := logRes["request_id"]
				if !ok {
					t.Errorf("expected a request id, got nothing")
					return
				}
				if val == "" {
					t.Errorf("expected a request id, got nothing")
					return
				}
			}
		})
	}
}
