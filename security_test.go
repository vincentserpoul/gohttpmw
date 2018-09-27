package gohttpmw

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurity(t *testing.T) {
	fakeHandler := http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {},
	)
	midWared := Security()(fakeHandler)
	rr := httptest.NewRecorder()
	request, errR := http.NewRequest("GET", ``, nil)
	if errR != nil {
		t.Fatalf("request creation failed %v", errR)
	}

	midWared.ServeHTTP(rr, request)
	if rr.Header().Get("X-Frame-Options") == "" {
		t.Errorf("expected %s, got nothing", "SAMEORIGIN")
	}
	if rr.Header().Get("X-Content-Type-Options") == "" {
		t.Errorf("expected %s, got nothing", "nosniff")
	}
	if rr.Header().Get("X-XSS-Protection") == "" {
		t.Errorf("expected %s, got nothing", "1; mode=block")
	}
	if rr.Header().Get("Referrer-Policy") == "" {
		t.Errorf("expected %s, got nothing", "same-origin")
	}
}
