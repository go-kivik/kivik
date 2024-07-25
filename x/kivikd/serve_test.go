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

package kivikd

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/go-kivik/kivik/v4/x/kivikd/conf"
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
