package client

import (
	"sort"
	"strings"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Config", configTest)
}

func configTest(clients *kt.Clients, conf kt.SuiteConfig, t *testing.T) {
	conf.Skip(t)
	clients.RunRW(t, func(t *testing.T) {
		conf.Skip(t)
		configRW(clients, conf, t)
	})
	clients.RunAdmin(t, func(t *testing.T) {
		conf.Skip(t)
		testConfig(clients.Admin, conf, t)
	})
	clients.RunNoAuth(t, func(t *testing.T) {
		conf.Skip(t)
		testConfig(clients.NoAuth, conf, t)
	})
}

func configRW(clients *kt.Clients, conf kt.SuiteConfig, t *testing.T) {
	clients.RunAdmin(t, func(t *testing.T) {
		conf.Skip(t)
		t.Run("Set", func(t *testing.T) {
			conf.Skip(t)
			testSet(clients.Admin, conf, t)
		})
		t.Run("Delete", func(t *testing.T) {
			conf.Skip(t)
			testDelete(clients.Admin, clients.Admin, conf, t)
		})
	})
	clients.RunNoAuth(t, func(t *testing.T) {
		conf.Skip(t)
		t.Run("Set", func(t *testing.T) {
			testSet(clients.NoAuth, conf, t)
		})
		t.Run("Delete", func(t *testing.T) {
			conf.Skip(t)
			testDelete(clients.Admin, clients.NoAuth, conf, t)
		})
	})
}

func testSet(client *kivik.Client, conf kt.SuiteConfig, t *testing.T) {
	c, _ := client.Config()
	defer c.Delete("kivik", "kivik")
	status := conf.Int(t, "status")
	err := c.Set("kivik", "kivik", "kivik")
	if kt.IsError(err, status, t) {
		return
	}
	if status > 0 {
		return
	}
	// Set should be 100% idempotent, so check that we get the same result
	err2 := c.Set("kivik", "kivik", "kivik")
	if errors.StatusCode(err) != errors.StatusCode(err2) {
		t.Errorf("Resetting config resulted in a different error. %s followed by %s", err, err2)
		return
	}
	t.Run("Retreive", func(t *testing.T) {
		conf.Skip(t)
		status := conf.Int(t, "status")
		value, err := c.Get("kivik", "kivik")
		if kt.IsError(err, status, t) {
			return
		}
		if status > 0 {
			return
		}
		if value != "kivik" {
			t.Errorf("Stored 'kivik', but retrieved '%s'", value)
		}
	})
}

func testDelete(admin, client *kivik.Client, conf kt.SuiteConfig, t *testing.T) {
	ac, _ := admin.Config()
	c, _ := client.Config()
	_ = ac.Set("kivik", "foo", "bar")
	defer ac.Delete("kivik", "foo")
	t.Run("NonExistantSection", func(t *testing.T) {
		kt.IsError(c.Delete("kivikkivik", "xyz"), conf.Int(t, "status"), t)
	})
	t.Run("NonExistantKey", func(t *testing.T) {
		kt.IsError(c.Delete("kivik", "bar"), conf.Int(t, "status"), t)
	})
	t.Run("ExistingKey", func(t *testing.T) {
		kt.IsError(c.Delete("kivik", "foo"), conf.Int(t, "status"), t)
	})
}

func testConfig(client *kivik.Client, conf kt.SuiteConfig, t *testing.T) {
	var c *config.Config
	{
		var err error
		c, err = client.Config()
		status := conf.Int(t, "status")
		if kt.IsError(err, status, t) {
			return
		}
		if status > 0 {
			return
		}
	}
	t.Run("GetAll", func(t *testing.T) {
		conf.Skip(t)
		t.Parallel()
		status := conf.Int(t, "status")
		all, err := c.GetAll()
		_ = kt.IsError(err, status, t)
		if status > 0 {
			return
		}
		sections := make([]string, 0, len(all))
		for sec := range all {
			sections = append(sections, sec)
		}
		sort.Strings(sections)
		if d := diff.TextSlices(conf.StringSlice(t, "expected_sections"), sections); d != "" {
			t.Errorf("GetAll() returned unexpected sections:\n%s\n", d)
		}
	})
	t.Run("GetSection", func(t *testing.T) {
		conf.Skip(t)
		t.Parallel()
		for _, secName := range conf.StringSlice(t, "sections") {
			func(secName string) {
				t.Run(secName, func(t *testing.T) {
					conf.Skip(t)
					t.Parallel()
					status := conf.Int(t, "status")
					sec, err := c.GetSection(secName)
					if kt.IsError(err, status, t) {
						return
					}
					if status > 0 {
						return
					}
					keys := make([]string, 0, len(sec))
					for key := range sec {
						keys = append(keys, key)
					}
					sort.Strings(keys)
					if d := diff.TextSlices(conf.StringSlice(t, "keys"), keys); d != "" {
						t.Errorf("GetSection() returned unexpected keys:\n%s\n", d)
					}
				})
			}(secName)
		}
	})
	t.Run("GetItem", func(t *testing.T) {
		conf.Skip(t)
		for _, item := range conf.StringSlice(t, "items") {
			func(item string) {
				i := strings.Split(item, ".")
				secName, key := i[0], i[1]
				t.Run(item, func(t *testing.T) {
					conf.Skip(t)
					t.Parallel()
					status := conf.Int(t, "status")
					value, err := c.Get(secName, key)
					if kt.IsError(err, status, t) {
						return
					}
					if status > 0 {
						return
					}
					expected := conf.String(t, "expected")
					if value != expected {
						t.Errorf("%s = '%s', expected '%s'", item, value, expected)
					}
				})
			}(item)
		}
	})
}
