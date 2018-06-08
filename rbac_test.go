package gohttpmw

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ory/ladon"
	manager "github.com/ory/ladon/manager/memory"
	"github.com/segmentio/ksuid"
)

func TestRBAC(t *testing.T) {
	fakeHandler := http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {},
	)

	warden := &ladon.Ladon{
		Manager: manager.NewMemoryManager(),
	}
	_ = warden.Manager.Create(
		&ladon.DefaultPolicy{
			ID:          ksuid.New().String(),
			Description: "Test GET",
			Subjects:    []string{"admin"},
			Resources: []string{
				"https://pol.com/test",
			},
			Actions: []string{"GET"},
			Effect:  ladon.AllowAccess,
		},
	)

	midWared := RBAC(warden, getRole)(fakeHandler)

	tests := []struct {
		name    string
		role    string
		method  string
		url     string
		allowed bool
	}{
		{
			name:    "allowed request",
			role:    "admin",
			method:  "GET",
			url:     "https://pol.com/test",
			allowed: true,
		},
		{
			name:    "not allowed request",
			role:    "admin",
			method:  "GET",
			url:     "https://pol.com",
			allowed: false,
		},
		{
			name:    "wrong role, not allowed",
			role:    "pollux",
			method:  "GET",
			url:     "https://pol.com/test",
			allowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), contextKeyRole, tt.role)
			request := httptest.NewRequest(tt.method, tt.url, nil)
			rr := httptest.NewRecorder()

			midWared.ServeHTTP(rr, request.WithContext(ctx))
			if tt.allowed && (rr.Code == http.StatusForbidden) ||
				!tt.allowed && (rr.Code != http.StatusForbidden) {
				t.Errorf("expected allowed %t, got %d", tt.allowed, rr.Code)
			}
		})
	}
}

func BenchmarkRBAC(b *testing.B) {
	fakeHandler := http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {},
	)

	warden := &ladon.Ladon{
		Manager: manager.NewMemoryManager(),
	}

	// 1000 policies
	for i := 0; i <= 1000; i++ {
		_ = warden.Manager.Create(
			&ladon.DefaultPolicy{
				ID:          ksuid.New().String(),
				Description: "Test GET",
				Subjects:    []string{"admin"},
				Resources: []string{
					"https://pol.com/test/" + string(i),
				},
				Actions: []string{"GET"},
				Effect:  ladon.AllowAccess,
			},
		)
	}

	midWared := RBAC(warden, getRole)(fakeHandler)

	ctx := context.WithValue(context.Background(), contextKeyRole, "admin")
	request := httptest.NewRequest("GET", "https://pol.com/test/245", nil)

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		midWared.ServeHTTP(rr, request.WithContext(ctx))
		resp := rr.Result()
		defer func() { _ = resp.Body.Close() }()
	}
}

const (
	contextKeyRole = ContextKey("role")
)

// getRole will retrieve the request id from the context if the user is an admin
func getRole(ctx context.Context) string {
	if role, ok := ctx.Value(contextKeyRole).(string); ok {
		return role
	}

	return ""
}
