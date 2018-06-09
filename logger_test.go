package gohttpmw

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestGetRequestError(t *testing.T) {
	var reqErr error

	ctx := context.Background()
	reqErr = GetRequestError(ctx)
	if reqErr != nil {
		t.Errorf("expected nothing, got %s", reqErr)
	}

	ctx = context.WithValue(ctx, ContextKeyRequestError, "testString")
	reqErr = GetRequestError(ctx)
	if reqErr != nil {
		t.Errorf("expected nothing, got %s", reqErr)
	}

	requestError := fmt.Errorf("test error")
	ctx = context.WithValue(ctx, ContextKeyRequestError, requestError)
	reqErr = GetRequestError(ctx)
	if reqErr != requestError {
		t.Errorf("expected %s, got %s", requestError, reqErr)
	}
}

func TestLoggerCompleteness(t *testing.T) {
	expectedNonEmptyFields := []string{
		"http_scheme",
		"http_proto",
		"http_method",
		"remote_addr",
		"user_agent",
		"host",
		"uri",
		"process_time",
		"http_status",
		"resp_length",
	}
	errTest := errors.New("test error")

	tc := []struct {
		name                   string
		handler                http.HandlerFunc
		expectedNonEmptyFields []string
		expectedLogLevel       zapcore.Level
		expectedHTTPStatus     int
		withHTTPS              bool
	}{
		{
			name: "classic request log",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			expectedNonEmptyFields: expectedNonEmptyFields,
			expectedLogLevel:       zap.InfoLevel,
			expectedHTTPStatus:     http.StatusOK,
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
			expectedNonEmptyFields: expectedNonEmptyFields,
			expectedLogLevel:       zap.ErrorLevel,
			expectedHTTPStatus:     http.StatusInternalServerError,
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
			expectedNonEmptyFields: expectedNonEmptyFields,
			expectedLogLevel:       zap.WarnLevel,
			expectedHTTPStatus:     http.StatusBadRequest,
		},
		{
			name: "request log with requestID",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					*req = *req.WithContext(
						context.WithValue(
							req.Context(),
							ContextKeyRequestID,
							ksuid.New(),
						),
					)
					w.WriteHeader(http.StatusOK)
				}),
			expectedNonEmptyFields: expectedNonEmptyFields,
			expectedLogLevel:       zap.InfoLevel,
			expectedHTTPStatus:     http.StatusOK,
		},
		{
			name: "404 request log",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}),
			expectedNonEmptyFields: expectedNonEmptyFields,
			expectedLogLevel:       zap.InfoLevel,
			expectedHTTPStatus:     http.StatusNotFound,
		},
		{
			name: "400 request log",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				}),
			expectedNonEmptyFields: expectedNonEmptyFields,
			expectedLogLevel:       zap.InfoLevel,
			expectedHTTPStatus:     http.StatusBadRequest,
		},
		{
			name: "https request log",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			expectedLogLevel:   zap.InfoLevel,
			withHTTPS:          true,
			expectedHTTPStatus: http.StatusOK,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			logger, observedLogs := observer.New(zapcore.InfoLevel)
			defer logger.Sync() // flushes buffer, if any
			sugar := zap.New(logger).Sugar()
			midWared := Logger(sugar)(tt.handler)
			rr := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", ``, nil)
			r.RemoteAddr = "127.0.0.1"
			r.Header.Set("User-Agent", "test")
			if tt.withHTTPS {
				r.TLS = &tls.ConnectionState{}
			}
			midWared.ServeHTTP(rr, r)

			allLogs := observedLogs.All()
			if len(allLogs) != 1 {
				t.Errorf("got %d logs instead of 1", len(allLogs))
				return
			}

			if tt.expectedLogLevel != allLogs[0].Level {
				t.Errorf(
					"wrong log level, expected %v, got %v",
					tt.expectedLogLevel, allLogs[0].Level,
				)
				return
			}

			for _, f := range allLogs[0].Context {

				switch f.Key {
				case "scheme":
					if tt.withHTTPS && f.String != "https" {
						t.Errorf(
							"wrong http scheme detected, expected https, got %s",
							f.String,
						)
						return
					}
				case "http_status":
					if int64(tt.expectedHTTPStatus) != f.Integer {
						t.Errorf(
							"wrong http status, expected %d, got %d",
							tt.expectedHTTPStatus,
							f.Integer,
						)
						return
					}
				}

				for _, field := range tt.expectedNonEmptyFields {
					found := false
					for _, f := range allLogs[0].Context {
						if f.Key == field {
							found = true
						}
					}
					if !found {
						t.Errorf("missing field %s in log", field)
						return
					}

				}
			}
		})
	}
}

// func TestLoggerData(t *testing.T) {
// 	reqIDTest := ksuid.New()
// 	tc := []struct {
// 		name                string
// 		handler             http.HandlerFunc
// 		requestError        ksuid.KSUID
// 		RequestErrorMessage string
// 		expectedLen         int
// 	}{
// 		{
// 			name: "classic request log",
// 			handler: http.HandlerFunc(
// 				func(w http.ResponseWriter, req *http.Request) {
// 					w.WriteHeader(http.StatusOK)
// 				}),
// 			RequestErrorMessage: "",
// 		},
// 		{
// 			name:      "classic request log with request ID",
// 			requestID: reqIDTest,
// 			handler: http.HandlerFunc(
// 				func(w http.ResponseWriter, req *http.Request) {
// 					w.WriteHeader(http.StatusOK)
// 				}),
// 			RequestErrorMessage: "",
// 		},
// 		{
// 			name: "classic request log with some bytes written",
// 			handler: http.HandlerFunc(
// 				func(w http.ResponseWriter, req *http.Request) {
// 					_, _ = w.Write([]byte{0x00})
// 					w.WriteHeader(http.StatusOK)
// 				}),
// 			RequestErrorMessage: "",
// 			expectedLen:         1,
// 		},
// 		{
// 			name: "request log with error",
// 			handler: http.HandlerFunc(
// 				func(w http.ResponseWriter, req *http.Request) {
// 					*req = *req.WithContext(
// 						context.WithValue(
// 							req.Context(),
// 							ErrRequestContextKey,
// 							errors.New("test error big"),
// 						),
// 					)
// 					w.WriteHeader(http.StatusInternalServerError)
// 				}),
// 			RequestErrorMessage: "test error big",
// 		},
// 	}

// 	for _, tt := range tc {
// 		t.Run(tt.name, func(t *testing.T) {
// 			logger, hook := test.NewNullLogger()
// 			midWared := Logger(logger)(tt.handler)
// 			rr := httptest.NewRecorder()
// 			r, _ := http.NewRequest("GET", ``, nil)
// 			r.RemoteAddr = "127.0.0.1"
// 			r.Header.Set("User-Agent", "test")
// 			if tt.requestID != ksuid.Nil {
// 				*r = *r.WithContext(
// 					context.WithValue(
// 						r.Context(),
// 						ContextKeyRequestID,
// 						reqIDTest,
// 					),
// 				)
// 			}
// 			midWared.ServeHTTP(rr, r)

// 			if len(hook.Entries) > 1 {
// 				t.Errorf("got %d logs instead of 1", len(hook.Entries))
// 				return
// 			}

// 			if tt.requestID != ksuid.Nil &&
// 				hook.LastEntry().Data["request_id"] != tt.requestID {
// 				t.Errorf(
// 					"missing requestid, expected %s, got %s",
// 					tt.requestID, hook.LastEntry().Data["request_id"])
// 				return
// 			}

// 			if string(hook.LastEntry().Message) != tt.RequestErrorMessage {
// 				t.Errorf(
// 					"missing error message `%s` in log, got `%s`",
// 					tt.RequestErrorMessage, hook.LastEntry().Message)
// 				return
// 			}

// 			if tt.expectedLen != hook.LastEntry().Data["resp_length"] {
// 				t.Errorf(
// 					"expected length %d got %d",
// 					tt.expectedLen, hook.LastEntry().Data["resp_length"])
// 				return
// 			}

// 		})
// 	}
// }

func Test_augmentedResponseWriter_Write(t *testing.T) {
	rr := httptest.NewRecorder()
	arw := newAugmentedResponseWriter(rr)
	testS := []byte("test")
	_, _ = arw.Write(testS)
	if arw.length != len(testS) {
		t.Errorf("augmentedResponseWriter.Write() length %d instead of %d",
			arw.length, len(testS),
		)
		return
	}

}
