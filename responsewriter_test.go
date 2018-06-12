package gohttpmw

import (
	"net/http/httptest"
	"testing"
)

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
