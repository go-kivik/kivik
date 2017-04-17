package serve

import (
	"context"
	"testing"
)

func TestBind(t *testing.T) {
	s := &Service{}
	if err := s.Bind(":9000"); err != nil {
		t.Errorf("Failed to parse ':9000': %s", err)
	}
	if host := s.Config().GetString(context.Background(), "httpd", "bind_address"); host != "" {
		t.Errorf("Host is '%s', expected ''", host)
	}
	if port := s.Config().GetInt(context.Background(), "httpd", "port"); port != 9000 {
		t.Errorf("Port is '%d', expected '9000'", port)
	}
}
