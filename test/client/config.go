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
	ctx.Run("group", func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.Run("Set", func(ctx *kt.Context) {
				testSet(ctx, ctx.Admin, "kivikadmin")
			})
			ctx.Run("Delete", func(ctx *kt.Context) {
				testDelete(ctx, ctx.Admin, "kivikadmin")
			})
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			ctx.Parallel()
			ctx.Run("Set", func(ctx *kt.Context) {
				testSet(ctx, ctx.NoAuth, "kiviknoauth")
			})
			ctx.Run("Delete", func(ctx *kt.Context) {
				testDelete(ctx, ctx.NoAuth, "kiviknoauth")
			})
		})
	})
}

func testSet(ctx *kt.Context, client *kivik.Client, name string) {
	ctx.Parallel()
	c, _ := client.Config()
	defer c.Delete(name, name)
	err := c.Set(name, name, name)
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	// Set should be 100% idempotent, so check that we get the same result
	err2 := c.Set(name, name, name+name)
	if errors.StatusCode(err) != errors.StatusCode(err2) {
		ctx.Errorf("Resetting config resulted in a different error. %s followed by %s", err, err2)
		return
	}
	ctx.Run("Retreive", func(ctx *kt.Context) {
		value, err := c.Get(name, name)
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		if value != name+name {
			ctx.Errorf("Stored '%s', but retrieved '%s'", name+name, value)
		}
	})
}

func testDelete(ctx *kt.Context, client *kivik.Client, secName string) {
	ctx.Parallel()
	ac, _ := ctx.Admin.Config()
	c, _ := client.Config()
	_ = ac.Set(secName, "foo", "bar")
	defer ac.Delete(secName, "foo")
	ctx.Run("NonExistantSection", func(ctx *kt.Context) {
		ctx.CheckError(c.Delete(secName+"nonexistant", "xyz"))
	})
	ctx.Run("NonExistantKey", func(ctx *kt.Context) {
		ctx.CheckError(c.Delete(secName, "baz"))
	})
	ctx.Run("ExistingKey", func(ctx *kt.Context) {
		ctx.CheckError(c.Delete(secName, "foo"))
	})
}

func testConfig(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	var c *config.Config
	{
		var err error
		c, err = client.Config()
		if !ctx.IsExpectedSuccess(err) {
			return
		}
	}
	ctx.Run("GetAll", func(ctx *kt.Context) {
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
		for _, s := range ctx.StringSlice("sections") {
			func(secName string) {
				ctx.Run(secName, func(ctx *kt.Context) {
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
		for _, i := range ctx.StringSlice("items") {
			func(item string) {
				ctx.Run(item, func(ctx *kt.Context) {
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
