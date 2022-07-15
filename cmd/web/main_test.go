package main

import "testing"

// TestRun checks if run function in main.go returns an error or not.
func TestRun(t *testing.T) {
	_, err := run()

	if err != nil {
		t.Error("failed run()")
	}
}
