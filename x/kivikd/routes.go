package kivikd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/go-kivik/kivikd/v4/couchserver"
	"github.com/go-kivik/kivikd/v4/logger"
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
