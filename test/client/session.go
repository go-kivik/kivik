package client

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/flimzy/kivik/driver/couchdb/chttp"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Session", session)
}

func session(ctx *kt.Context) {
	ctx.Run("Get", func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testSession(ctx, ctx.CHTTPAdmin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testSession(ctx, ctx.CHTTPNoAuth)
		})
	})
}

func testSession(ctx *kt.Context, client *chttp.Client) {
	if client == nil {
		ctx.Skipf("No CHTTP client")
	}
	uCtx := struct {
		Info struct {
			AuthMethod   string   `json:"authenticated"`
			AuthDB       string   `json:"authentication_db"`
			AuthHandlers []string `json:"authentication_handlers"`
		} `json:"info"`
		OK      bool `json:"ok"`
		UserCtx struct {
			Name  string   `json:"name"`
			Roles []string `json:"roles"`
		} `json:"userCtx"`
	}{}
	err := client.DoJSON(kt.CTX, chttp.MethodGet, "/_session", nil, &uCtx)
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	values := map[string]string{
		"info.authenticated":           uCtx.Info.AuthMethod,
		"info.authentication_db":       uCtx.Info.AuthDB,
		"info.authentication_handlers": strings.Join(uCtx.Info.AuthHandlers, ","),
		"ok":            fmt.Sprintf("%t", uCtx.OK),
		"userCtx.roles": strings.Join(uCtx.UserCtx.Roles, ","),
	}
	for key, actual := range values {
		expected := ctx.MustString(key)
		if actual != expected {
			ctx.Errorf("Unexpected value for `%s`. Expected '%s', actual '%s'", key, expected, actual)
		}
	}
	dsn, _ := url.Parse(client.DSN())
	var expected string
	if dsn.User != nil {
		expected = dsn.User.Username()
	}
	actual := uCtx.UserCtx.Name
	if actual != expected {
		ctx.Errorf("Unexpected value for `%s`. Expected '%s', actual '%s'", "userCtx.name", expected, actual)
	}
}
