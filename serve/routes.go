package serve

import (
	"fmt"
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler"
	"github.com/dimfeld/httptreemux"
	"github.com/justinas/alice"
)

func (s *Service) setupRoutes() (http.Handler, error) {
	router := httptreemux.New()
	router.HeadCanUseGet = true
	ctxRoot := router.UsingContext()
	ctxRoot.Handler(mGET, "/", handler(root))
	ctxRoot.Handler(mGET, "/favicon.ico", handler(favicon))
	ctxRoot.Handler(mGET, "/_all_dbs", handler(allDBs))
	ctxRoot.Handler(mPUT, "/:db", handler(createDB))
	ctxRoot.Handler(mHEAD, "/:db", handler(dbExists))
	ctxRoot.Handler(mPOST, "/:db/_ensure_full_commit", handler(flush))

	ctxRoot.Handler(mGET, "/_session", handler(getSession))
	// Note that DELETE and POST for the /_session endpoint are handled by the
	// cookie auth handler. This means if you aren't using cookie auth, that
	// these methods will return 405.

	// ctxRoot.Handler(mDELETE, "/:db", handler(destroyDB) )
	// ctxRoot.Handler(http.MethodGet, "/:db", handler(getDB))

	return alice.New(
		setContext(s),
		setSession(),
		requestLogger,
		gzipHandler(s),
		authHandler,
	).Then(router), nil
}

func gzipHandler(s *Service) func(http.Handler) http.Handler {
	level := s.Conf().GetInt("httpd.compression_level")
	if level == 0 {
		level = 8
	}
	gzipHandler, err := gziphandler.NewGzipLevelHandler(int(level))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid httpd.compression_level %d\n", level)
		return func(h http.Handler) http.Handler {
			return h
		}
	}
	fmt.Fprintf(os.Stderr, "Enabling HTTPD cmpression, level %d\n", level)
	return gzipHandler
}
