package conf

import (
	"strings"
	"testing"
)

const testConfDir = "../test/conf"

func TestLoadError(t *testing.T) {
	_, err := Load(testConfDir + "/serve.invalid")
	if err == nil || err.Error() != `Unsupported Config Type "invalid"` {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestLoad(t *testing.T) {
	c, err := Load(testConfDir + "/serve.toml")
	if err != nil {
		t.Errorf("Failed to load config: %s", err)
	}
	if v := c.GetString("httpd.bind_address"); v != "0.0.0.0" {
		t.Errorf("Unexpected value %s", v)
	}
}

func TestLoadDefault(t *testing.T) {
	_, err := Load("")
	if err != nil && !strings.HasPrefix(err.Error(), `Config File "serve" Not Found in`) {
		t.Errorf("Failed to load default config: %s", err)
	}
}
