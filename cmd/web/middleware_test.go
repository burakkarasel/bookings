package main

import (
	"fmt"
	"net/http"
	"testing"
)

// TestNoSurf is test func for NoSurf func in middleware.go. It checks return type of the func NoSurf.
func TestNoSurf(t *testing.T) {
	var myH myHandler

	h := NoSurf(&myH)

	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, it is %T", v))
	}
}

// TestSessionLoad is test func for SessionLoad func in routes.go. It checks return type of the func SessionLoad.
func TestSessionLoad(t *testing.T) {
	var myH myHandler

	h := SessionLoad(&myH)

	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, it is %T", v))
	}
}
