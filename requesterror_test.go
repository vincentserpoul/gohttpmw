package gohttpmw

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
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

func TestSetRequestError(t *testing.T) {
	err := fmt.Errorf("test")
	req := httptest.NewRequest("GET", "/", nil)
	SetRequestError(req, err)
	if GetRequestError(req.Context()) != err {
		t.Errorf("SetRequestError didn't set the error in the context")
		return
	}
}
