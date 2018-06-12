package gohttpmw

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestLogger(t *testing.T) {
	expectedNonEmptyFields := []string{"http_scheme",
		"http_proto", "http_method", "remote_addr", "user_agent", "uri",
		"process_time", "http_status", "resp_length"}
	errTest := errors.New("test error")

	tc := []struct {
		name                   string
		handler                http.HandlerFunc
		expectedNonEmptyFields []string
		expectedLogLevel       logrus.Level
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
			expectedLogLevel:       logrus.InfoLevel,
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
			expectedLogLevel:       logrus.ErrorLevel,
			expectedHTTPStatus:     http.StatusInternalServerError,
		},
		{
			name: "404 request log",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}),
			expectedNonEmptyFields: expectedNonEmptyFields,
			expectedLogLevel:       logrus.InfoLevel,
			expectedHTTPStatus:     http.StatusNotFound,
		},
		{
			name: "400 request log",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				}),
			expectedNonEmptyFields: expectedNonEmptyFields,
			expectedLogLevel:       logrus.InfoLevel,
			expectedHTTPStatus:     http.StatusBadRequest,
		},
		{
			name: "400 request log with error",
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
			expectedLogLevel:       logrus.WarnLevel,
			expectedHTTPStatus:     http.StatusBadRequest,
		},
		{
			name: "https request log",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			expectedLogLevel:   logrus.InfoLevel,
			withHTTPS:          true,
			expectedHTTPStatus: http.StatusOK,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			logger, hook := test.NewNullLogger()
			midWared := Logger(logger)(tt.handler)
			rr := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, `/`, nil)
			r.RemoteAddr = "127.0.0.1"
			r.Header.Set("User-Agent", "test")
			if tt.withHTTPS {
				r.TLS = &tls.ConnectionState{}
			}
			midWared.ServeHTTP(rr, r)

			if len(hook.Entries) > 1 {
				t.Errorf("got %d logs instead of 1", len(hook.Entries))
				return
			}

			if tt.expectedLogLevel != hook.LastEntry().Level {
				t.Errorf(
					"wrong log level, expected %v, got %v",
					tt.expectedLogLevel, hook.LastEntry().Level,
				)
				return
			}

			if tt.withHTTPS && hook.LastEntry().Data["http_scheme"] != "https" {
				t.Errorf(
					"wrong http scheme detected, expected https, got %s",
					hook.LastEntry().Data["http_scheme"],
				)
				return
			}

			if hook.LastEntry().Data["http_status"] != tt.expectedHTTPStatus {
				t.Errorf(
					"wrong http status, expected %d, got %d",
					tt.expectedHTTPStatus, hook.LastEntry().Data["http_status"],
				)
				return
			}

			for _, field := range tt.expectedNonEmptyFields {
				if hook.LastEntry().Data[field] == nil {
					t.Errorf("missing field %s in log", field)
					return
				}
			}
		})
	}
}

func TestLoggerData(t *testing.T) {
	reqIDTest := xid.New()
	tc := []struct {
		name                string
		handler             http.HandlerFunc
		requestID           xid.ID
		requestErrorMessage string
		withRequestAdd      bool
		expectedLen         int
	}{
		{
			name: "classic request log",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			requestErrorMessage: "",
		},
		{
			name: "classic request log with request ID",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					*req = *req.WithContext(
						context.WithValue(
							req.Context(),
							ContextKeyRequestID,
							reqIDTest.String(),
						),
					)
				}),
			requestErrorMessage: "",
			requestID:           reqIDTest,
		},
		{
			name: "classic request log with some bytes written",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					_, _ = w.Write([]byte{0x00})
				}),
			requestErrorMessage: "",
			expectedLen:         1,
		},
		{
			name: "request log with error",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					*req = *req.WithContext(
						context.WithValue(
							req.Context(),
							ContextKeyRequestError,
							errors.New("test error big"),
						),
					)
					w.WriteHeader(http.StatusInternalServerError)
				}),
			requestErrorMessage: "test error big",
		},
		{
			name: "request log with additional fields",
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {
					*req = *req.WithContext(
						context.WithValue(
							req.Context(),
							ContextKeyAddToRequestLog,
							map[string]interface{}{"fish": "fish"},
						),
					)
				}),
			withRequestAdd: true,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			logger, hook := test.NewNullLogger()
			midWared := Logger(logger)(tt.handler)
			rr := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, `/`, nil)
			r.RemoteAddr = "127.0.0.1"
			r.Header.Set("User-Agent", "test")
			midWared.ServeHTTP(rr, r)

			if len(hook.Entries) > 1 {
				t.Errorf("got %d logs instead of 1", len(hook.Entries))
				return
			}

			if tt.requestID != (xid.ID{}) {
				val, ok := hook.LastEntry().Data["request_id"]
				if !ok {
					t.Errorf("missing requestid")
					return
				}
				if val != tt.requestID.String() {
					t.Errorf(
						"missing requestid, expected %s, got %s",
						tt.requestID, val)
					return
				}
			}

			if string(hook.LastEntry().Message) != tt.requestErrorMessage {
				t.Errorf(
					"missing error message `%s` in log, got `%s`",
					tt.requestErrorMessage, hook.LastEntry().Message)
				return
			}

			if tt.expectedLen != hook.LastEntry().Data["resp_length"] {
				t.Errorf(
					"expected length %d got %d",
					tt.expectedLen, hook.LastEntry().Data["resp_length"])
				return
			}

			if tt.withRequestAdd {
				if hook.LastEntry().Data["fish"] != "fish" {
					t.Errorf("expected additional field but none present")
					return
				}
			}

		})
	}
}

func TestAddToRequestLog(t *testing.T) {
	fFish := func(ctx context.Context) interface{} { return "fish" }
	fNil := func(ctx context.Context) interface{} { return nil }
	tc := []struct {
		name          string
		f             func(context.Context) interface{}
		expectedKey   string
		expectedValue interface{}
	}{
		{
			name:          "f returns a value",
			f:             fFish,
			expectedKey:   "fish",
			expectedValue: "fish",
		},
		{
			name:          "f returns no value",
			f:             fNil,
			expectedKey:   "fish",
			expectedValue: nil,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			fakeHandler := http.HandlerFunc(
				func(w http.ResponseWriter, req *http.Request) {},
			)
			midWared := AddToRequestLog(tt.expectedKey, tt.f)(fakeHandler)
			request := httptest.NewRequest(http.MethodGet, `/`, nil)

			midWared.ServeHTTP(nil, request)
			ct := GetAddToRequestLog(request.Context())
			if tt.expectedValue == nil {
				if _, ok := ct[tt.expectedKey]; ok {
					t.Errorf(
						"AddToRequestLog didn't added an empty field",
					)
					return
				}
			} else if !reflect.DeepEqual(
				ct,
				map[string]interface{}{tt.expectedKey: tt.expectedValue},
			) {
				t.Errorf("AddToRequestLog didn't add the field to the context")
				return
			}
		})
	}
}

func TestGetAddToRequestLog(t *testing.T) {
	ctx := context.Background()
	if !reflect.DeepEqual(
		GetAddToRequestLog(ctx),
		map[string]interface{}{},
	) {
		t.Errorf("AddToRequestLog didn't add the field to the context")
		return
	}

	testATL := map[string]interface{}{"fish": "fish"}
	if atl := GetAddToRequestLog(
		context.WithValue(ctx, ContextKeyAddToRequestLog, testATL),
	); !reflect.DeepEqual(atl, testATL) {
		t.Errorf("expected %v, got %v", testATL, atl)
		return
	}
}
