package gohttpmw

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCSP(t *testing.T) {
	fakeHandler := http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {},
	)
	midWared := CSP("test")(fakeHandler)
	rr := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, `/`, nil)

	midWared.ServeHTTP(rr, request)
	if rr.Header().Get("content-security-policy") == "" {
		t.Errorf("expected a csp, got nothing")
	}
}
