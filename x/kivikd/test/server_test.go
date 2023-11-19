//go:build !js
// +build !js

package test

import (
	"testing"

	_ "github.com/go-kivik/kivik/v4/x/kivikd"
)

func init() {
	RegisterKivikdSuites()
}

func TestServer(t *testing.T) {
	ServerTest(t)
}
