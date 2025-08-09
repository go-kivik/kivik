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

package mockdb

import (
	"context"
	"errors"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

type mockTest struct {
	setup func(*Client)
	test  func(*testing.T, *kivik.Client)
	err   string
}

func testMock(t *testing.T, test mockTest) {
	t.Helper()
	client, mock, err := New()
	if err != nil {
		t.Fatalf("error creating mock database: %s", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})
	if test.setup != nil {
		test.setup(mock)
	}
	if test.test != nil {
		test.test(t, client)
	}
	err = mock.ExpectationsWereMet()
	if !testy.ErrorMatchesRE(test.err, err) {
		t.Errorf("ExpectationsWereMet returned unexpected error: %s", err)
	}
}

func TestCloseClient(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("err", mockTest{
		setup: func(m *Client) {
			m.ExpectClose().WillReturnError(errors.New("close failed"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // not a helper
			err := c.Close()
			if !testy.ErrorMatches("close failed", err) {
				t.Errorf("unexpected error: %s", err)
			}
		},
		err: "",
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // not a helper
			err := c.Close()
			const want = "call to Close() was not expected, all expectations already fulfilled"
			if !testy.ErrorMatches(want, err) {
				t.Errorf("unexpected error: %s", err)
			}
		},
	})
	tests.Add("callback", mockTest{
		setup: func(m *Client) {
			m.ExpectClose().WillExecute(func() error {
				return errors.New("custom error")
			})
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.Close()
			if !testy.ErrorMatches("custom error", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestAllDBs(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectAllDBs().WillReturnError(errors.New("AllDBs failed"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.AllDBs(context.TODO())
			if !testy.ErrorMatches("AllDBs failed", err) {
				t.Errorf("unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.AllDBs(context.TODO())
			if !testy.ErrorMatches("call to AllDBs() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("success", func() interface{} {
		expected := []string{"a", "b", "c"}
		return mockTest{
			setup: func(m *Client) {
				m.ExpectAllDBs().WillReturn(expected)
			},
			test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
				result, err := c.AllDBs(context.TODO())
				if !testy.ErrorMatches("", err) {
					t.Errorf("Unexpected error: %s", err)
				}
				if d := testy.DiffInterface(expected, result); d != nil {
					t.Error(d)
				}
			},
		}
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectAllDBs().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.AllDBs(newCanceledContext())
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("options", mockTest{
		setup: func(m *Client) {
			m.ExpectAllDBs().WithOptions(kivik.Param("foo", 123))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.AllDBs(context.TODO(), kivik.Param("foo", 123))
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("callback", mockTest{
		setup: func(m *Client) {
			m.ExpectAllDBs().WillExecute(func(context.Context, driver.Options) ([]string, error) {
				return nil, errors.New("custom error")
			})
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.AllDBs(context.TODO())
			if !testy.ErrorMatches("custom error", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestClusterSetup(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectClusterSetup().WillReturnError(errors.New("setup error"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.ClusterSetup(context.TODO(), 123)
			if !testy.ErrorMatches("setup error", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("action", mockTest{
		setup: func(m *Client) {
			m.ExpectClusterSetup().WithAction(123)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.ClusterSetup(context.TODO(), 123)
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectClusterSetup().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.ClusterSetup(newCanceledContext(), 123)
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.ClusterSetup(context.TODO(), 123)
			if !testy.ErrorMatches("call to ClusterSetup() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("callback", mockTest{
		setup: func(m *Client) {
			m.ExpectClusterSetup().WillExecute(func(context.Context, interface{}) error {
				return errors.New("custom error")
			})
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.ClusterSetup(context.TODO(), 123)
			if !testy.ErrorMatches("custom error", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestClusterStatus(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectClusterStatus().WillReturnError(errors.New("status error"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.ClusterStatus(context.TODO())
			if !testy.ErrorMatches("status error", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("options", mockTest{
		setup: func(m *Client) {
			m.ExpectClusterStatus().WithOptions(kivik.Param("foo", 123))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.ClusterStatus(context.TODO())
			if !testy.ErrorMatchesRE(`map\[foo:123]`, err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
		err: "there is a remaining unmet expectation",
	})
	tests.Add("success", func() interface{} {
		const expected = "oink"
		return mockTest{
			setup: func(m *Client) {
				m.ExpectClusterStatus().WillReturn(expected)
			},
			test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
				result, err := c.ClusterStatus(context.TODO())
				if !testy.ErrorMatches("", err) {
					t.Errorf("Unexpected error: %s", err)
				}
				if result != expected {
					t.Errorf("Unexpected result: %s", result)
				}
			},
		}
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectClusterStatus().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.ClusterStatus(newCanceledContext())
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unordered", mockTest{
		setup: func(m *Client) {
			m.ExpectClose()
			m.ExpectClusterStatus()
			m.MatchExpectationsInOrder(false)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.ClusterStatus(context.TODO())
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
		err: "there is a remaining unmet expectation: call to Close()",
	})
	tests.Add("unexpected", mockTest{
		setup: func(m *Client) {
			m.ExpectClose()
			m.MatchExpectationsInOrder(false)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.ClusterStatus(context.TODO())
			if !testy.ErrorMatches("call to ClusterStatus(ctx, [?]) was not expected", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
		err: "there is a remaining unmet expectation: call to Close()",
	})
	tests.Add("callback", mockTest{
		setup: func(m *Client) {
			m.ExpectClusterStatus().WillExecute(func(context.Context, driver.Options) (string, error) {
				return "", errors.New("custom error")
			})
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.ClusterStatus(newCanceledContext())
			if !testy.ErrorMatches("custom error", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestDBExists(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectDBExists().WillReturnError(errors.New("existence error"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.DBExists(context.TODO(), "foo")
			if !testy.ErrorMatches("existence error", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("name", mockTest{
		setup: func(m *Client) {
			m.ExpectDBExists().WithName("foo")
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			exists, err := c.DBExists(context.TODO(), "foo")
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if exists {
				t.Errorf("DB shouldn't exist")
			}
		},
	})
	tests.Add("options", mockTest{
		setup: func(m *Client) {
			m.ExpectDBExists().WithOptions(kivik.Param("foo", 123))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.DBExists(context.TODO(), "foo")
			if !testy.ErrorMatchesRE(`map\[foo:123]`, err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
		err: "there is a remaining unmet expectation",
	})
	tests.Add("exists", mockTest{
		setup: func(m *Client) {
			m.ExpectDBExists().WillReturn(true)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			exists, err := c.DBExists(context.TODO(), "foo")
			if !testy.ErrorMatchesRE("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if !exists {
				t.Errorf("DB should exist")
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectDBExists().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.DBExists(newCanceledContext(), "foo")
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestDestroyDB(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectDestroyDB().WillReturnError(errors.New("foo err"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.DestroyDB(newCanceledContext(), "foo")
			if !testy.ErrorMatches("foo err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("name", mockTest{
		setup: func(m *Client) {
			m.ExpectDestroyDB().WithName("foo")
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.DestroyDB(newCanceledContext(), "foo")
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("options", mockTest{
		setup: func(m *Client) {
			m.ExpectDestroyDB().WithOptions(kivik.Param("foo", 123))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.DestroyDB(newCanceledContext(), "foo")
			if !testy.ErrorMatchesRE(`map\[foo:123]`, err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
		err: "there is a remaining unmet expectation",
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectDestroyDB().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.DestroyDB(newCanceledContext(), "foo")
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestDBsStats(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectDBsStats().WillReturnError(errors.New("stats error"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.DBsStats(context.TODO(), []string{"foo"})
			if !testy.ErrorMatches("stats error", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("names", mockTest{
		setup: func(m *Client) {
			m.ExpectDBsStats().WithNames([]string{"a", "b"})
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.DBsStats(context.TODO(), []string{"foo"})
			if !testy.ErrorMatchesRE("[a b]", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
		err: "there is a remaining unmet expectation",
	})
	tests.Add("success", func() interface{} {
		return mockTest{
			setup: func(m *Client) {
				m.ExpectDBsStats().WillReturn([]*driver.DBStats{
					{Name: "foo", Cluster: &driver.ClusterStats{Replicas: 5}},
					{Name: "bar"},
				})
			},
			test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
				result, err := c.DBsStats(context.TODO(), []string{"foo", "bar"})
				if !testy.ErrorMatchesRE("", err) {
					t.Errorf("Unexpected error: %s", err)
				}
				expected := []*kivik.DBStats{
					{Name: "foo", Cluster: &kivik.ClusterConfig{Replicas: 5}},
					{Name: "bar"},
				}
				if d := testy.DiffInterface(expected, result); d != nil {
					t.Error(d)
				}
			},
		}
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectDBsStats().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.DBsStats(newCanceledContext(), []string{"foo"})
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestPing(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("unreachable", mockTest{
		setup: func(m *Client) {
			m.ExpectPing()
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			reachable, err := c.Ping(context.TODO())
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if reachable {
				t.Errorf("Expected db to be unreachable")
			}
		},
	})
	tests.Add("reachable", mockTest{
		setup: func(m *Client) {
			m.ExpectPing().WillReturn(true)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			reachable, err := c.Ping(context.TODO())
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if !reachable {
				t.Errorf("Expected db to be reachable")
			}
		},
	})
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectPing().WillReturnError(errors.New("foo err"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Ping(context.TODO())
			if !testy.ErrorMatches("foo err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Ping(context.TODO())
			if !testy.ErrorMatches("call to Ping() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectPing().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Ping(newCanceledContext())
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestSession(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("session", func() interface{} {
		return mockTest{
			setup: func(m *Client) {
				m.ExpectSession().WillReturn(&driver.Session{
					Name: "bob",
				})
			},
			test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
				session, err := c.Session(context.TODO())
				if !testy.ErrorMatches("", err) {
					t.Errorf("Unexpected error: %s", err)
				}
				expected := &kivik.Session{
					Name: "bob",
				}
				if d := testy.DiffInterface(expected, session); d != nil {
					t.Error(d)
				}
			},
		}
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Session(context.TODO())
			if !testy.ErrorMatches("call to Session() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectSession().WillReturnError(errors.New("foo err"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Session(context.TODO())
			if !testy.ErrorMatches("foo err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectSession().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Session(newCanceledContext())
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestVersion(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("version", func() interface{} {
		return mockTest{
			setup: func(m *Client) {
				m.ExpectVersion().WillReturn(&driver.Version{Version: "1.2"})
			},
			test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
				session, err := c.Version(context.TODO())
				if !testy.ErrorMatches("", err) {
					t.Errorf("Unexpected error: %s", err)
				}
				expected := &kivik.ServerVersion{Version: "1.2"}
				if d := testy.DiffInterface(expected, session); d != nil {
					t.Error(d)
				}
			},
		}
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Version(context.TODO())
			if !testy.ErrorMatches("call to Version() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectVersion().WillReturnError(errors.New("foo err"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Version(context.TODO())
			if !testy.ErrorMatches("foo err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectVersion().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Version(newCanceledContext())
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestDB(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("name", mockTest{
		setup: func(m *Client) {
			m.ExpectDB().WithName("foo")
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.DB("foo").Err()
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.DB("foo").Err()
			if !testy.ErrorMatches("call to DB() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("options", mockTest{
		setup: func(m *Client) {
			m.ExpectDB().WithOptions(kivik.Param("foo", 123))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.DB("foo", kivik.Param("foo", 123)).Err()
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("success", mockTest{
		setup: func(m *Client) {
			m.ExpectDB().WillReturn(m.NewDB())
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			db := c.DB("asd")
			err := db.Err()
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if db.Name() != "asd" {
				t.Errorf("Unexpected db name: %s", db.Name())
			}
		},
	})
	tests.Run(t, testMock)
}

func TestCreateDB(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectCreateDB().WillReturnError(errors.New("foo err"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.CreateDB(context.TODO(), "foo")
			if !testy.ErrorMatches("foo err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("name", mockTest{
		setup: func(m *Client) {
			m.ExpectCreateDB().WithName("foo")
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.CreateDB(context.TODO(), "foo")
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.CreateDB(context.TODO(), "foo")
			if !testy.ErrorMatches("call to CreateDB() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("options", mockTest{
		setup: func(m *Client) {
			m.ExpectCreateDB().WithOptions(kivik.Param("foo", 123))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.CreateDB(context.TODO(), "foo", kivik.Param("foo", 123))
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("success", mockTest{
		setup: func(m *Client) {
			m.ExpectCreateDB()
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.CreateDB(context.TODO(), "foo")
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectCreateDB().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.CreateDB(newCanceledContext(), "foo")
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("cleanup expectations", mockTest{
		setup: func(m *Client) {
			m.ExpectCreateDB().WillReturnError(errors.New("foo err"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.CreateDB(context.TODO(), "foo")
			if err == nil {
				t.Fatal("expected error")
			}
		},
	})
	tests.Add("callback", mockTest{
		setup: func(m *Client) {
			m.ExpectCreateDB().WillExecute(func(context.Context, string, driver.Options) error {
				return errors.New("custom error")
			})
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			err := c.CreateDB(context.TODO(), "foo")
			if !testy.ErrorMatches("custom error", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestDBUpdates(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectDBUpdates().WillReturnError(errors.New("foo err"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			rows := c.DBUpdates(context.TODO())
			if err := rows.Err(); !testy.ErrorMatches("foo err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			rows := c.DBUpdates(context.TODO())
			if err := rows.Err(); !testy.ErrorMatches("call to DBUpdates() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("close error", mockTest{
		setup: func(m *Client) {
			m.ExpectDBUpdates().WillReturn(NewDBUpdates().CloseError(errors.New("bar err")))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			rows := c.DBUpdates(context.TODO())
			if err := rows.Err(); !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if err := rows.Close(); !testy.ErrorMatches("bar err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("updates", mockTest{
		setup: func(m *Client) {
			m.ExpectDBUpdates().WillReturn(NewDBUpdates().
				AddUpdate(&driver.DBUpdate{DBName: "foo"}).
				AddUpdate(&driver.DBUpdate{DBName: "bar"}).
				AddUpdate(&driver.DBUpdate{DBName: "baz"}))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			rows := c.DBUpdates(context.TODO())
			if err := rows.Err(); !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
			names := []string{}
			for rows.Next() {
				names = append(names, rows.DBName())
			}
			expected := []string{"foo", "bar", "baz"}
			if d := testy.DiffInterface(expected, names); d != nil {
				t.Error(d)
			}
		},
	})
	tests.Add("iter error", mockTest{
		setup: func(m *Client) {
			m.ExpectDBUpdates().WillReturn(NewDBUpdates().
				AddUpdate(&driver.DBUpdate{DBName: "foo"}).
				AddUpdateError(errors.New("foo err")))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			rows := c.DBUpdates(context.TODO())
			if err := rows.Err(); !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
			names := []string{}
			for rows.Next() {
				names = append(names, rows.DBName())
			}
			expected := []string{"foo"}
			if d := testy.DiffInterface(expected, names); d != nil {
				t.Error(d)
			}
			if err := rows.Err(); !testy.ErrorMatches("foo err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectDBUpdates().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			rows := c.DBUpdates(newCanceledContext())
			if err := rows.Err(); !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("update delay", mockTest{
		setup: func(m *Client) {
			m.ExpectDBUpdates().WillReturn(NewDBUpdates().
				AddDelay(time.Millisecond).
				AddUpdate(&driver.DBUpdate{DBName: "foo"}).
				AddDelay(time.Second).
				AddUpdate(&driver.DBUpdate{DBName: "bar"}))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			rows := c.DBUpdates(ctx)
			if err := rows.Err(); !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
			names := []string{}
			for rows.Next() {
				names = append(names, rows.DBName())
			}
			expected := []string{"foo"}
			if d := testy.DiffInterface(expected, names); d != nil {
				t.Error(d)
			}
			if err := rows.Err(); !testy.ErrorMatches("context deadline exceeded", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestConfig(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectConfig().WillReturnError(errors.New("foo err"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Config(context.TODO(), "local")
			if !testy.ErrorMatches("foo err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Config(context.TODO(), "local")
			if !testy.ErrorMatches("call to Config() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectConfig().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Config(newCanceledContext(), "local")
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("success", func() interface{} {
		expected := kivik.Config{"foo": kivik.ConfigSection{"bar": "baz"}}
		return mockTest{
			setup: func(m *Client) {
				m.ExpectConfig().
					WithNode("local").
					WillReturn(driver.Config{"foo": driver.ConfigSection{"bar": "baz"}})
			},
			test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
				result, err := c.Config(newCanceledContext(), "local")
				if !testy.ErrorMatches("", err) {
					t.Errorf("Unexpected error: %s", err)
				}
				if d := testy.DiffInterface(expected, result); d != nil {
					t.Error(d)
				}
			},
		}
	})

	tests.Run(t, testMock)
}

func TestConfigSection(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectConfigSection().WillReturnError(errors.New("foo err"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.ConfigSection(context.TODO(), "local", "foo")
			if !testy.ErrorMatches("foo err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.ConfigSection(context.TODO(), "local", "foo")
			if !testy.ErrorMatches("call to ConfigSection() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectConfigSection().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.ConfigSection(newCanceledContext(), "local", "foo")
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("success", func() interface{} {
		expected := kivik.ConfigSection{"bar": "baz"}
		return mockTest{
			setup: func(m *Client) {
				m.ExpectConfigSection().
					WithNode("local").
					WithSection("foo").
					WillReturn(driver.ConfigSection{"bar": "baz"})
			},
			test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
				result, err := c.ConfigSection(newCanceledContext(), "local", "foo")
				if !testy.ErrorMatches("", err) {
					t.Errorf("Unexpected error: %s", err)
				}
				if d := testy.DiffInterface(expected, result); d != nil {
					t.Error(d)
				}
			},
		}
	})

	tests.Run(t, testMock)
}

func TestConfigValue(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectConfigValue().WillReturnError(errors.New("foo err"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.ConfigValue(context.TODO(), "local", "foo", "bar")
			if !testy.ErrorMatches("foo err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.ConfigValue(context.TODO(), "local", "foo", "bar")
			if !testy.ErrorMatches("call to ConfigValue() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectConfigValue().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.ConfigValue(newCanceledContext(), "local", "foo", "bar")
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("success", func() interface{} {
		expected := "baz"
		return mockTest{
			setup: func(m *Client) {
				m.ExpectConfigValue().
					WithNode("local").
					WithSection("foo").
					WithKey("bar").
					WillReturn("baz")
			},
			test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
				result, err := c.ConfigValue(newCanceledContext(), "local", "foo", "bar")
				if !testy.ErrorMatches("", err) {
					t.Errorf("Unexpected error: %s", err)
				}
				if d := testy.DiffInterface(expected, result); d != nil {
					t.Error(d)
				}
			},
		}
	})

	tests.Run(t, testMock)
}

func TestSetConfigValue(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectSetConfigValue().WillReturnError(errors.New("foo err"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.SetConfigValue(context.TODO(), "local", "foo", "bar", "baz")
			if !testy.ErrorMatches("foo err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.SetConfigValue(context.TODO(), "local", "foo", "bar", "baz")
			if !testy.ErrorMatches("call to SetConfigValue() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectSetConfigValue().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.SetConfigValue(newCanceledContext(), "local", "foo", "bar", "baz")
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("success", func() interface{} {
		expected := "old"
		return mockTest{
			setup: func(m *Client) {
				m.ExpectSetConfigValue().
					WithNode("local").
					WithSection("foo").
					WithKey("bar").
					WithValue("baz").
					WillReturn("old")
			},
			test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
				result, err := c.SetConfigValue(newCanceledContext(), "local", "foo", "bar", "baz")
				if !testy.ErrorMatches("", err) {
					t.Errorf("Unexpected error: %s", err)
				}
				if d := testy.DiffInterface(expected, result); d != nil {
					t.Error(d)
				}
			},
		}
	})

	tests.Run(t, testMock)
}

func TestDeleteConfigKey(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("error", mockTest{
		setup: func(m *Client) {
			m.ExpectDeleteConfigKey().WillReturnError(errors.New("foo err"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.DeleteConfigKey(context.TODO(), "local", "foo", "bar")
			if !testy.ErrorMatches("foo err", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.DeleteConfigKey(context.TODO(), "local", "foo", "bar")
			if !testy.ErrorMatches("call to DeleteConfigKey() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectDeleteConfigKey().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.DeleteConfigKey(newCanceledContext(), "local", "foo", "bar")
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("success", func() interface{} {
		expected := "old"
		return mockTest{
			setup: func(m *Client) {
				m.ExpectDeleteConfigKey().
					WithNode("local").
					WithSection("foo").
					WithKey("bar").
					WillReturn("old")
			},
			test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
				result, err := c.DeleteConfigKey(newCanceledContext(), "local", "foo", "bar")
				if !testy.ErrorMatches("", err) {
					t.Errorf("Unexpected error: %s", err)
				}
				if d := testy.DiffInterface(expected, result); d != nil {
					t.Error(d)
				}
			},
		}
	})

	tests.Run(t, testMock)
}

func TestReplicate(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("err", mockTest{
		setup: func(m *Client) {
			m.ExpectReplicate().WillReturnError(errors.New("replicate failed"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Replicate(context.TODO(), "foo", "bar")
			if !testy.ErrorMatches("replicate failed", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Replicate(context.TODO(), "foo", "bar")
			if !testy.ErrorMatches("call to Replicate() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("source and target", mockTest{
		setup: func(m *Client) {
			m.ExpectReplicate().
				WithSource("bar").
				WithTarget("foo").
				WillReturnError(errors.New("expected"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Replicate(context.TODO(), "foo", "bar")
			if !testy.ErrorMatches("expected", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("return", mockTest{
		setup: func(m *Client) {
			r := m.NewReplication().ID("aaa")
			m.ExpectReplicate().
				WillReturn(r)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			rep, err := c.Replicate(context.TODO(), "foo", "bar")
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if id := rep.ReplicationID(); id != "aaa" {
				t.Errorf("Unexpected replication ID: %s", id)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectReplicate().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.Replicate(newCanceledContext(), "foo", "bar")
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}

func TestGetReplications(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("err", mockTest{
		setup: func(m *Client) {
			m.ExpectGetReplications().WillReturnError(errors.New("get replications failed"))
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.GetReplications(context.TODO())
			if !testy.ErrorMatches("get replications failed", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("unexpected", mockTest{
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.GetReplications(context.TODO())
			if !testy.ErrorMatches("call to GetReplications() was not expected, all expectations already fulfilled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Add("return", mockTest{
		setup: func(m *Client) {
			m.ExpectGetReplications().
				WillReturn([]*Replication{
					m.NewReplication().ID("bbb"),
					m.NewReplication().ID("ccc"),
				})
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			reps, err := c.GetReplications(context.TODO())
			if !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if id := reps[0].ReplicationID(); id != "bbb" {
				t.Errorf("Unexpected replication 1 ID: %s", id)
			}
			if id := reps[1].ReplicationID(); id != "ccc" {
				t.Errorf("Unexpected replication 2 ID: %s", id)
			}
		},
	})
	tests.Add("delay", mockTest{
		setup: func(m *Client) {
			m.ExpectGetReplications().WillDelay(time.Second)
		},
		test: func(t *testing.T, c *kivik.Client) { //nolint:thelper // Not a helper
			_, err := c.GetReplications(newCanceledContext())
			if !testy.ErrorMatches("context canceled", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		},
	})
	tests.Run(t, testMock)
}
