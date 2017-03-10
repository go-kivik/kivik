package client

import (
	"bytes"
	"strings"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Log", log)
}

type logTest struct {
	name     string
	len      int64
	offset   int64
	contains string
}

func log(ctx *kt.Context) {
	// Do a request that will be easy to find in the logs
	ctx.Admin.DBExists("abracadabra")
	tests := []logTest{
		logTest{
			name: "Len0",
		},
		logTest{
			name: "Len1000",
			len:  1000,
		},
		logTest{
			name:     "Contains",
			len:      10000,
			contains: "HEAD /abracadabra",
		},
		logTest{
			name:   "Offset1000",
			len:    1000,
			offset: 1000,
		},
		logTest{
			name:   "Offset-1000",
			len:    1000,
			offset: -1000,
		},
	}
	ctx.RunAdmin(func(ctx *kt.Context) {
		for _, test := range tests {
			doLogTest(ctx, ctx.Admin, test)
		}
		rawLogTests(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		for _, test := range tests {
			doLogTest(ctx, ctx.NoAuth, test)
		}
		rawLogTests(ctx, ctx.NoAuth)
	})
}

func doLogTest(ctx *kt.Context, client *kivik.Client, test logTest) {
	ctx.Run(test.name, func(ctx *kt.Context) {
		ctx.Parallel()
		logR, err := client.Log(test.len, test.offset)
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		buf := &bytes.Buffer{}
		buf.ReadFrom(logR)
		log := buf.String()
		if test.contains != "" && !strings.Contains(log, test.contains) {
			ctx.Errorf("Log does not contain expected string '%s':\n%s", test.contains, log)
		}
	})
}

func rawLogTests(ctx *kt.Context, client *kivik.Client) {
	ctx.Run("HTTP", func(ctx *kt.Context) {
		rawLogTest(ctx, client, "NegativeBytes", "/_log?bytes=-1")
		rawLogTest(ctx, client, "TextBytes", "/_log?bytes=chicken")
		rawLogTest(ctx, client, "BogusQueryParam", "/_log?chicken=yummy")
	})
}

func rawLogTest(ctx *kt.Context, client *kivik.Client, name, path string) {
	ctx.Run(name, func(ctx *kt.Context) {
		ctx.Parallel()
		// req, httpClient, err := client.HTTPRequest("GET", path, nil)
		// if err != nil {
		// 	ctx.Fatalf("Error creating request: %s", err)
		// }
		// resp, err := httpClient.Do(req)
		// if err != nil {
		// 	ctx.Errorf("Failed to send HTTP request: %s", err)
		// 	return
		// }
		// ctx.CheckError(couchdb.ResponseError(resp))
	})
}
