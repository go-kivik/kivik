package serve

import (
	"testing"

	"github.com/flimzy/kivik/serve/conf"
	"github.com/spf13/viper"
)

func TestBind(t *testing.T) {
	s := &Service{
		Config: &conf.Conf{Viper: viper.New()},
	}
	if err := s.Bind(":9000"); err != nil {
		t.Errorf("Failed to parse ':9000': %s", err)
	}
	if host := s.Conf().GetString("httpd.bind_address"); host != "" {
		t.Errorf("Host is '%s', expected ''", host)
	}
	if port := s.Conf().GetInt("httpd.port"); port != 9000 {
		t.Errorf("Port is '%d', expected '9000'", port)
	}
}
