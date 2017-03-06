package client

import (
	"sort"
	"strings"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Config", configTest)
}

func configTest(ctx *kt.Context) {
	ctx.RunRW(func(ctx *kt.Context) {
		configRW(ctx)
	})
	ctx.RunAdmin(func(ctx *kt.Context) {
		testConfig(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testConfig(ctx, ctx.NoAuth)
	})
}

func configRW(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		ctx.Run("Set", func(ctx *kt.Context) {
			testSet(ctx, ctx.Admin)
		})
		ctx.Run("Delete", func(ctx *kt.Context) {
			testDelete(ctx, ctx.Admin)
		})
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		ctx.Run("Set", func(ctx *kt.Context) {
			testSet(ctx, ctx.NoAuth)
		})
		ctx.Run("Delete", func(ctx *kt.Context) {
			testDelete(ctx, ctx.NoAuth)
		})
	})
}

func testSet(ctx *kt.Context, client *kivik.Client) {
	c, _ := client.Config()
	defer c.Delete("kivik", "kivik")
	err := c.Set("kivik", "kivik", "kivik")
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	// Set should be 100% idempotent, so check that we get the same result
	err2 := c.Set("kivik", "kivik", "kivik")
	if errors.StatusCode(err) != errors.StatusCode(err2) {
		ctx.Errorf("Resetting config resulted in a different error. %s followed by %s", err, err2)
		return
	}
	ctx.Run("Retreive", func(ctx *kt.Context) {
		value, err := c.Get("kivik", "kivik")
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		if value != "kivik" {
			ctx.Errorf("Stored 'kivik', but retrieved '%s'", value)
		}
	})
}

func testDelete(ctx *kt.Context, client *kivik.Client) {
	ac, _ := ctx.Admin.Config()
	c, _ := client.Config()
	_ = ac.Set("kivik", "foo", "bar")
	defer ac.Delete("kivik", "foo")
	ctx.Run("NonExistantSection", func(ctx *kt.Context) {
		ctx.CheckError(c.Delete("kivikkivik", "xyz"))
	})
	ctx.Run("NonExistantKey", func(ctx *kt.Context) {
		ctx.CheckError(c.Delete("kivik", "bar"))
	})
	ctx.Run("ExistingKey", func(ctx *kt.Context) {
		ctx.CheckError(c.Delete("kivik", "foo"))
	})
}

func testConfig(ctx *kt.Context, client *kivik.Client) {
	var c *config.Config
	{
		var err error
		c, err = client.Config()
		if !ctx.IsExpectedSuccess(err) {
			return
		}
	}
	ctx.Run("GetAll", func(ctx *kt.Context) {
		ctx.Parallel()
		all, err := c.GetAll()
		if !ctx.IsSuccess(err) {
			return
		}
		sections := make([]string, 0, len(all))
		for sec := range all {
			sections = append(sections, sec)
		}
		sort.Strings(sections)
		if d := diff.TextSlices(ctx.StringSlice("expected_sections"), sections); d != "" {
			ctx.Errorf("GetAll() returned unexpected sections:\n%s\n", d)
		}
	})
	ctx.Run("GetSection", func(ctx *kt.Context) {
		ctx.Parallel()
		for _, s := range ctx.StringSlice("sections") {
			func(secName string) {
				ctx.Run(secName, func(ctx *kt.Context) {
					ctx.Parallel()
					sec, err := c.GetSection(secName)
					if !ctx.IsExpectedSuccess(err) {
						return
					}
					keys := make([]string, 0, len(sec))
					for key := range sec {
						keys = append(keys, key)
					}
					sort.Strings(keys)
					if d := diff.TextSlices(ctx.StringSlice("keys"), keys); d != "" {
						ctx.Errorf("GetSection() returned unexpected keys:\n%s\n", d)
					}
				})
			}(s)
		}
	})
	ctx.Run("GetItem", func(ctx *kt.Context) {
		ctx.Parallel()
		for _, i := range ctx.StringSlice("items") {
			func(item string) {
				ctx.Run(item, func(ctx *kt.Context) {
					ctx.Parallel()
					parts := strings.Split(item, ".")
					secName, key := parts[0], parts[1]
					value, err := c.Get(secName, key)
					if !ctx.IsExpectedSuccess(err) {
						return
					}
					expected := ctx.String("expected")
					if value != expected {
						ctx.Errorf("%s = '%s', expected '%s'", item, value, expected)
					}
				})
			}(i)
		}
	})
}
