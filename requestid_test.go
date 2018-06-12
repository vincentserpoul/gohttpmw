package gohttpmw

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/xid"
)

func TestRequestID(t *testing.T) {
	fakeHandler := http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {},
	)
	midWared := RequestID()(fakeHandler)
	rr := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, `/`, nil)

	midWared.ServeHTTP(rr, request)
	if rr.Header().Get("requestID") == "" {
		t.Errorf("expected a requestid, got nothing")
	}
}

func TestGetRequestID(t *testing.T) {
	ctx := context.Background()
	if reqID := GetRequestID(ctx); reqID != "" {
		t.Errorf("expected nothing, got %s", reqID)
		return
	}

	testReq := xid.New().String()
	if reqID := GetRequestID(
		context.WithValue(ctx, ContextKeyRequestID, testReq),
	); reqID != testReq {
		t.Errorf("expected %s, got %s", testReq, reqID)
		return
	}
}
