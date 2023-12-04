// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

//go:build !js
// +build !js

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
