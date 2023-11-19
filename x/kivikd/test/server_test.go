//go:build !js
// +build !js

package test

import (
	"testing"

	_ "github.com/go-kivik/kivikd/v4"
)

func init() {
	RegisterKivikdSuites()
}

func TestServer(t *testing.T) {
	ServerTest(t)
}
