package client

import (
	"sort"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	kt.Register("Config", config)
}

func config(clients *kt.Clients, conf kt.SuiteConfig, t *testing.T) {
	clients.RunRW(t, func(t *testing.T) {
		configRW(clients, conf, t)
	})
	clients.RunAdmin(t, func(t *testing.T) {
		testConfig(clients.Admin, conf, t)
	})
}

func configRW(clients *kt.Clients, conf kt.SuiteConfig, t *testing.T) {
}

func testConfig(client *kivik.Client, conf kt.SuiteConfig, t *testing.T) {
	c, err := client.Config()
	status := conf.Int(t, "status")
	_ = kt.IsError(err, status, t)
	if status > 0 {
		return
	}
	t.Run("GetAll", func(t *testing.T) {
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
			t.Errorf("Config() returned unexpected sections:\n%s\n", d)
		}

	})

}
