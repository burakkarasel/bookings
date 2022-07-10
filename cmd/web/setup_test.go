package main

import (
	"net/http"
	"os"
	"testing"
)

// myHandler struct implements ServeHTTP interface
type myHandler struct {
}

// TestMain starts running before our other test files run, and it exits after our tests completed
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// ServeHTTP is created to implement handler interface to myHandler struct
func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// do nothing
}
