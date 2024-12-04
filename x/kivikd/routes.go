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
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-kivik/kivik/v4/x/kivikd/couchserver"
	"github.com/go-kivik/kivik/v4/x/kivikd/logger"
)

func (s *Service) setupRoutes() (http.Handler, error) {
	h := couchserver.NewHandler(s.Client)
	h.Vendor = s.VendorName
	h.VendorVersion = s.VendorVersion
	h.Favicon = s.Favicon
	h.SessionKey = SessionKey

	rlog := s.RequestLogger
	if rlog == nil {
		rlog = logger.DefaultLogger
	}

	return chi.Chain(
		setContext(s),
		setSession(),
		loggerMiddleware(rlog),
		gzipHandler(s),
		authHandler,
	).Handler(h.Main()), nil
}

func gzipHandler(s *Service) func(http.Handler) http.Handler {
	level := s.Conf().GetInt("httpd.compression_level")
	if level == 0 {
		level = 8
	}
	fmt.Fprintf(os.Stderr, "Enabling gzip compression, level %d\n", level)
	return middleware.Compress(level, "text/plain", "text/html", "application/json")
}
