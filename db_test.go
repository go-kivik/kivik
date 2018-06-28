package kivik

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
)

func TestClient(t *testing.T) {
	client := &Client{}
	db := &DB{client: client}
	result := db.Client()
	if result != client {
		t.Errorf("Unexpected result. Expected %p, got %p", client, result)
	}
}

func TestName(t *testing.T) {
	dbName := "foo"
	db := &DB{name: dbName}
	result := db.Name()
	if result != dbName {
		t.Errorf("Unexpected result. Expected %s, got %s", dbName, result)
	}
}

func TestAllDocs(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		options  Options
		expected *Rows
		status   int
		err      string
	}{
		{
			name: "db error",
			db: &DB{
				driverDB: &mockDB{
					AllDocsFunc: func(_ context.Context, _ map[string]interface{}) (driver.Rows, error) {
						return nil, errors.New("db error")
					},
				},
			},
			status: StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mockDB{
					AllDocsFunc: func(_ context.Context, opts map[string]interface{}) (driver.Rows, error) {
						if d := diff.Interface(testOptions, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options: %s", d)
						}
						return &mockRows{id: "a"}, nil
					},
				},
			},
			options: testOptions,
			expected: &Rows{
				iter: &iter{
					feed: &rowsIterator{
						Rows: &mockRows{id: "a"},
					},
					curVal: &driver.Row{},
				},
				rowsi: &mockRows{id: "a"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.AllDocs(context.Background(), test.options)
			testy.StatusError(t, test.err, test.status, err)
			result.cancel = nil // Determinism
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestQuery(t *testing.T) {
	tests := []struct {
		name       string
		db         *DB
		ddoc, view string
		options    Options
		expected   *Rows
		status     int
		err        string
	}{
		{
			name: "db error",
			db: &DB{
				driverDB: &mockDB{
					QueryFunc: func(_ context.Context, ddoc, view string, opts map[string]interface{}) (driver.Rows, error) {
						return nil, errors.New("db error")
					},
				},
			},
			status: StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mockDB{
					QueryFunc: func(_ context.Context, ddoc, view string, opts map[string]interface{}) (driver.Rows, error) {
						expectedDdoc := "foo"
						expectedView := "bar"
						if ddoc != expectedDdoc {
							return nil, fmt.Errorf("Unexpected ddoc: %s", ddoc)
						}
						if view != expectedView {
							return nil, fmt.Errorf("Unexpected view: %s", view)
						}
						if d := diff.Interface(testOptions, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options: %s", d)
						}
						return &mockRows{id: "a"}, nil
					},
				},
			},
			ddoc:    "foo",
			view:    "bar",
			options: testOptions,
			expected: &Rows{
				iter: &iter{
					feed: &rowsIterator{
						Rows: &mockRows{id: "a"},
					},
					curVal: &driver.Row{},
				},
				rowsi: &mockRows{id: "a"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Query(context.Background(), test.ddoc, test.view, test.options)
			testy.StatusError(t, test.err, test.status, err)
			result.cancel = nil // Determinism
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		docID    string
		options  Options
		expected *Row
		status   int
		err      string
	}{
		{
			name: "db error",
			db: &DB{
				driverDB: &mockDB{
					GetFunc: func(_ context.Context, _ string, _ map[string]interface{}) (json.RawMessage, error) {
						return nil, errors.New("db error")
					},
				},
			},
			status: StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mockDB{
					GetFunc: func(_ context.Context, docID string, options map[string]interface{}) (json.RawMessage, error) {
						expectedDocID := "foo"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						if d := diff.Interface(testOptions, options); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%s", d)
						}
						return []byte(`{"_id":"foo"}`), nil
					},
				},
			},
			docID:   "foo",
			options: testOptions,
			expected: &Row{
				doc: []byte(`{"_id":"foo"}`),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Get(context.Background(), test.docID, test.options)
			testy.StatusError(t, test.err, test.status, err)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestFlush(t *testing.T) {
	tests := []struct {
		name   string
		db     *DB
		status int
		err    string
	}{
		{
			name: "non-Flusher",
			db: &DB{
				driverDB: &mockDB{},
			},
			status: StatusNotImplemented,
			err:    "kivik: flush not supported by driver",
		},
		{
			name: "db error",
			db: &DB{
				driverDB: &mockDBFlusher{
					FlushFunc: func(_ context.Context) error {
						return errors.Status(StatusBadResponse, "flush error")
					},
				},
			},
			status: StatusBadResponse,
			err:    "flush error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mockDBFlusher{
					FlushFunc: func(_ context.Context) error {
						return nil
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.db.Flush(context.Background())
			testy.StatusError(t, test.err, test.status, err)
		})
	}
}

func TestStats(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		expected *DBStats
		status   int
		err      string
	}{
		{
			name: "stats error",
			db: &DB{
				driverDB: &mockDB{
					StatsFunc: func(_ context.Context) (*driver.DBStats, error) {
						return nil, errors.Status(StatusBadResponse, "stats error")
					},
				},
			},
			status: StatusBadResponse,
			err:    "stats error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mockDB{
					StatsFunc: func(_ context.Context) (*driver.DBStats, error) {
						return &driver.DBStats{
							Name:           "foo",
							CompactRunning: true,
							DocCount:       1,
							DeletedCount:   2,
							UpdateSeq:      "abc",
							DiskSize:       3,
							ActiveSize:     4,
							ExternalSize:   5,
							Cluster: &driver.ClusterStats{
								Replicas:    6,
								Shards:      7,
								ReadQuorum:  8,
								WriteQuorum: 9,
							},
							RawResponse: []byte("foo"),
						}, nil
					},
				},
			},
			expected: &DBStats{
				Name:           "foo",
				CompactRunning: true,
				DocCount:       1,
				DeletedCount:   2,
				UpdateSeq:      "abc",
				DiskSize:       3,
				ActiveSize:     4,
				ExternalSize:   5,
				Cluster: &ClusterConfig{
					Replicas:    6,
					Shards:      7,
					ReadQuorum:  8,
					WriteQuorum: 9,
				},
				RawResponse: []byte("foo"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Stats(context.Background())
			testy.StatusError(t, test.err, test.status, err)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestCompact(t *testing.T) {
	expected := "compact error"
	db := &DB{
		driverDB: &mockDB{
			CompactFunc: func(_ context.Context) error {
				return errors.Status(StatusBadRequest, expected)
			},
		},
	}
	err := db.Compact(context.Background())
	testy.StatusError(t, expected, StatusBadRequest, err)
}

func TestCompactView(t *testing.T) {
	expectedDDocID := "foo"
	expected := "compact view error"
	db := &DB{
		driverDB: &mockDB{
			CompactViewFunc: func(_ context.Context, ddocID string) error {
				if ddocID != expectedDDocID {
					return fmt.Errorf("Unexpected ddocID: %s", ddocID)
				}
				return errors.Status(StatusBadRequest, expected)
			},
		},
	}
	err := db.CompactView(context.Background(), expectedDDocID)
	testy.StatusError(t, expected, StatusBadRequest, err)
}

func TestViewCleanup(t *testing.T) {
	expected := "compact error"
	db := &DB{
		driverDB: &mockDB{
			ViewCleanupFunc: func(_ context.Context) error {
				return errors.Status(StatusBadRequest, expected)
			},
		},
	}
	err := db.ViewCleanup(context.Background())
	testy.StatusError(t, expected, StatusBadRequest, err)
}

func TestSecurity(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		expected *Security
		status   int
		err      string
	}{
		{
			name: "security error",
			db: &DB{
				driverDB: &mockDB{
					SecurityFunc: func(_ context.Context) (*driver.Security, error) {
						return nil, errors.Status(StatusBadResponse, "security error")
					},
				},
			},
			status: StatusBadResponse,
			err:    "security error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mockDB{
					SecurityFunc: func(_ context.Context) (*driver.Security, error) {
						return &driver.Security{
							Admins: driver.Members{
								Names: []string{"a"},
								Roles: []string{"b"},
							},
							Members: driver.Members{
								Names: []string{"c"},
								Roles: []string{"d"},
							},
						}, nil
					},
				},
			},
			expected: &Security{
				Admins: Members{
					Names: []string{"a"},
					Roles: []string{"b"},
				},
				Members: Members{
					Names: []string{"c"},
					Roles: []string{"d"},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Security(context.Background())
			testy.StatusError(t, test.err, test.status, err)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestSetSecurity(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		security *Security
		status   int
		err      string
	}{
		{
			name:   "nil security",
			status: StatusBadRequest,
			err:    "kivik: security required",
		},
		{
			name: "set error",
			db: &DB{
				driverDB: &mockDB{
					SetSecurityFunc: func(_ context.Context, _ *driver.Security) error {
						return errors.Status(StatusBadResponse, "set security error")
					},
				},
			},
			security: &Security{},
			status:   StatusBadResponse,
			err:      "set security error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mockDB{
					SetSecurityFunc: func(_ context.Context, security *driver.Security) error {
						expectedSecurity := &driver.Security{
							Admins: driver.Members{
								Names: []string{"a"},
								Roles: []string{"b"},
							},
							Members: driver.Members{
								Names: []string{"c"},
								Roles: []string{"d"},
							},
						}
						if d := diff.Interface(expectedSecurity, security); d != nil {
							return fmt.Errorf("Unexpected security:\n%s", d)
						}
						return nil
					},
				},
			},
			security: &Security{
				Admins: Members{
					Names: []string{"a"},
					Roles: []string{"b"},
				},
				Members: Members{
					Names: []string{"c"},
					Roles: []string{"d"},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.db.SetSecurity(context.Background(), test.security)
			testy.StatusError(t, test.err, test.status, err)
		})
	}
}

func TestRev(t *testing.T) {
	tests := []struct {
		name     string
		db       *DB
		docID    string
		expected string
		status   int
		err      string
	}{
		{
			name: "rever error",
			db: &DB{
				driverDB: &mockRever{
					RevFunc: func(_ context.Context, _ string) (string, error) {
						return "", errors.Status(StatusBadResponse, "rever error")
					},
				},
			},
			status: StatusBadResponse,
			err:    "rever error",
		},
		{
			name: "rever success",
			db: &DB{
				driverDB: &mockRever{
					RevFunc: func(_ context.Context, docID string) (string, error) {
						expectedDocID := "foo"
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						return "1-xxx", nil
					},
				},
			},
			docID:    "foo",
			expected: "1-xxx",
		},
		{
			name: "non-rever error",
			db: &DB{
				driverDB: &mockDB{
					GetFunc: func(_ context.Context, _ string, _ map[string]interface{}) (json.RawMessage, error) {
						return nil, errors.Status(StatusBadResponse, "get error")
					},
				},
			},
			status: StatusBadResponse,
			err:    "get error",
		},
		{
			name: "non-rever invalid json",
			db: &DB{
				driverDB: &mockDB{
					GetFunc: func(_ context.Context, _ string, _ map[string]interface{}) (json.RawMessage, error) {
						return []byte("invalid json"), nil
					},
				},
			},
			status: StatusBadResponse,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name: "non-rever success",
			db: &DB{
				driverDB: &mockDB{
					GetFunc: func(_ context.Context, docID string, opts map[string]interface{}) (json.RawMessage, error) {
						expectedDocID := "foo"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						if opts != nil {
							return nil, errors.New("opts should be nil")
						}
						return []byte(`{"_rev":"1-xxx"}`), nil
					},
				},
			},
			docID:    "foo",
			expected: "1-xxx",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Rev(context.Background(), test.docID)
			testy.StatusError(t, test.err, test.status, err)
			if result != test.expected {
				t.Errorf("Unexpected result: %s", result)
			}
		})
	}
}

func TestCopy(t *testing.T) {
	tests := []struct {
		name           string
		db             *DB
		target, source string
		options        Options
		expected       string
		status         int
		err            string
	}{
		{
			name:   "missing target",
			status: StatusBadRequest,
			err:    "kivik: targetID required",
		},
		{
			name:   "missing source",
			target: "foo",
			status: StatusBadRequest,
			err:    "kivik: sourceID required",
		},
		{
			name: "copier error",
			db: &DB{
				driverDB: &mockCopier{
					CopyFunc: func(_ context.Context, _, _ string, _ map[string]interface{}) (string, error) {
						return "", errors.Status(StatusBadRequest, "copy error")
					},
				},
			},
			target: "foo",
			source: "bar",
			status: StatusBadRequest,
			err:    "copy error",
		},
		{
			name: "copier success",
			db: &DB{
				driverDB: &mockCopier{
					CopyFunc: func(_ context.Context, target, source string, options map[string]interface{}) (string, error) {
						expectedTarget := "foo"
						expectedSource := "bar"
						if target != expectedTarget {
							return "", fmt.Errorf("Unexpected target: %s", target)
						}
						if source != expectedSource {
							return "", fmt.Errorf("Unexpected source: %s", source)
						}
						if d := diff.Interface(testOptions, options); d != nil {
							return "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "1-xxx", nil
					},
				},
			},
			target:   "foo",
			source:   "bar",
			options:  testOptions,
			expected: "1-xxx",
		},
		{
			name: "non-copier get error",
			db: &DB{
				driverDB: &mockDB{
					GetFunc: func(_ context.Context, _ string, _ map[string]interface{}) (json.RawMessage, error) {
						return nil, errors.Status(StatusBadResponse, "get error")
					},
				},
			},
			target: "foo",
			source: "bar",
			status: StatusBadResponse,
			err:    "get error",
		},
		{
			name: "non-copier invalid JSON",
			db: &DB{
				driverDB: &mockDB{
					GetFunc: func(_ context.Context, _ string, _ map[string]interface{}) (json.RawMessage, error) {
						return []byte("invalid json"), nil
					},
				},
			},
			target: "foo",
			source: "bar",
			status: StatusBadResponse,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name: "non-copier put error",
			db: &DB{
				driverDB: &mockDB{
					GetFunc: func(_ context.Context, _ string, _ map[string]interface{}) (json.RawMessage, error) {
						return []byte(`{"_id":"foo","_rev":"1-xxx"}`), nil
					},
					PutFunc: func(_ context.Context, _ string, _ interface{}) (string, error) {
						return "", errors.Status(StatusBadResponse, "put error")
					},
				},
			},
			target: "foo",
			source: "bar",
			status: StatusBadResponse,
			err:    "put error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mockDB{
					GetFunc: func(_ context.Context, docID string, options map[string]interface{}) (json.RawMessage, error) {
						expectedDocID := "bar"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected get docID: %s", docID)
						}
						return []byte(`{"_id":"bar","_rev":"1-xxx","foo":123.4}`), nil
					},
					PutFunc: func(_ context.Context, docID string, doc interface{}) (string, error) {
						expectedDocID := "foo"
						expectedDoc := map[string]interface{}{"_id": "foo", "foo": 123.4}
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected put docID: %s", docID)
						}
						if d := diff.Interface(expectedDoc, doc); d != nil {
							return "", fmt.Errorf("Unexpected doc:\n%s", doc)
						}
						return "1-xxx", nil
					},
				},
			},
			target:   "foo",
			source:   "bar",
			expected: "1-xxx",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Copy(context.Background(), test.target, test.source, test.options)
			testy.StatusError(t, test.err, test.status, err)
			if result != test.expected {
				t.Errorf("Unexpected result: %s", result)
			}
		})
	}
}

type errorReader struct{}

var _ io.Reader = &errorReader{}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("errorReader")
}

func TestNormalizeFromJSON(t *testing.T) {
	type njTest struct {
		Name     string
		Input    interface{}
		Expected interface{}
		Status   int
		Error    string
	}
	tests := []njTest{
		{
			Name:     "Interface",
			Input:    int(5),
			Expected: int(5),
		},
		{
			Name:   "InvalidJSON",
			Input:  []byte(`invalid`),
			Status: StatusBadRequest,
			Error:  "invalid character 'i' looking for beginning of value",
		},
		{
			Name:     "Bytes",
			Input:    []byte(`{"foo":"bar"}`),
			Expected: map[string]interface{}{"foo": "bar"},
		},
		{
			Name:     "RawMessage",
			Input:    json.RawMessage(`{"foo":"bar"}`),
			Expected: map[string]interface{}{"foo": "bar"},
		},
		{
			Name:     "ioReader",
			Input:    strings.NewReader(`{"foo":"bar"}`),
			Expected: map[string]interface{}{"foo": "bar"},
		},
		{
			Name:   "ErrorReader",
			Input:  &errorReader{},
			Status: StatusUnknownError,
			Error:  "errorReader",
		},
	}
	for _, test := range tests {
		func(test njTest) {
			t.Run(test.Name, func(t *testing.T) {
				result, err := normalizeFromJSON(test.Input)
				var msg string
				var status int
				if err != nil {
					msg = err.Error()
					status = StatusCode(err)
				}
				if msg != test.Error || status != test.Status {
					t.Errorf("Unexpected error: %d %s", status, msg)
				}
				if err != nil {
					return
				}
				if d := diff.Interface(test.Expected, result); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}

func TestPut(t *testing.T) {
	putOptsFunc := func(_ context.Context, docID string, doc interface{}, opts map[string]interface{}) (string, error) {
		expectedDocID := "foo"
		expectedDoc := map[string]interface{}{"foo": "bar"}
		if expectedDocID != docID {
			return "", errors.Errorf("Unexpected docID: %s", docID)
		}
		if d := diff.Interface(expectedDoc, doc); d != nil {
			return "", errors.Errorf("Unexpected doc: %s", d)
		}
		if d := diff.Interface(testOptions, opts); d != nil {
			return "", errors.Errorf("Unexpected opts: %s", d)
		}
		return "1-xxx", nil
	}
	type putTest struct {
		name    string
		db      *DB
		docID   string
		input   interface{}
		options Options
		status  int
		err     string
		newRev  string
	}
	tests := []putTest{
		{
			name:   "no docID",
			status: StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name: "db error",
			db: &DB{
				driverDB: &mockDBOpts{
					PutOptsFunc: func(_ context.Context, _ string, _ interface{}, _ map[string]interface{}) (string, error) {
						return "", errors.Status(StatusBadRequest, "db error")
					},
				},
			},
			docID:  "foo",
			status: StatusBadRequest,
			err:    "db error",
		},
		{
			name: "Interface",
			db: &DB{
				driverDB: &mockDBOpts{
					PutOptsFunc: putOptsFunc,
				},
			},
			docID:   "foo",
			input:   map[string]interface{}{"foo": "bar"},
			options: testOptions,
			newRev:  "1-xxx",
		},
		{
			name:   "InvalidJSON",
			docID:  "foo",
			input:  []byte("Something bogus"),
			status: StatusBadRequest,
			err:    "invalid character 'S' looking for beginning of value",
		},
		{
			name: "Bytes",
			db: &DB{
				driverDB: &mockDBOpts{
					PutOptsFunc: putOptsFunc,
				},
			},
			docID:   "foo",
			input:   []byte(`{"foo":"bar"}`),
			options: testOptions,
			newRev:  "1-xxx",
		},
		{
			name: "RawMessage",
			db: &DB{
				driverDB: &mockDBOpts{
					PutOptsFunc: putOptsFunc,
				},
			},
			docID:   "foo",
			input:   json.RawMessage(`{"foo":"bar"}`),
			options: testOptions,
			newRev:  "1-xxx",
		},
		{
			name: "Reader",
			db: &DB{
				driverDB: &mockDBOpts{
					PutOptsFunc: putOptsFunc,
				},
			},
			docID:   "foo",
			input:   strings.NewReader(`{"foo":"bar"}`),
			options: testOptions,
			newRev:  "1-xxx",
		},
		{
			name:   "ErrorReader",
			docID:  "foo",
			input:  &errorReader{},
			status: StatusUnknownError,
			err:    "errorReader",
		},
		{
			name: "legacy",
			db: &DB{
				driverDB: &mockDB{
					PutFunc: func(_ context.Context, docID string, doc interface{}) (string, error) {
						expectedDocID := "foo"
						expectedDoc := map[string]string{"foo": "bar"}
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						if d := diff.Interface(expectedDoc, doc); d != nil {
							return "", fmt.Errorf("Unexpected doc:\n%s", d)
						}
						return "1-xxx", nil
					},
				},
			},
			docID:  "foo",
			input:  map[string]string{"foo": "bar"},
			newRev: "1-xxx",
		},
	}
	for _, test := range tests {
		func(test putTest) {
			t.Run(test.name, func(t *testing.T) {
				newRev, err := test.db.Put(context.Background(), test.docID, test.input, test.options)
				testy.StatusError(t, test.err, test.status, err)
				if newRev != test.newRev {
					t.Errorf("Unexpected new rev: %s", newRev)
				}
			})
		}(test)
	}
}

func TestExtractDocID(t *testing.T) {
	type ediTest struct {
		name     string
		i        interface{}
		id       string
		expected bool
	}
	tests := []ediTest{
		{
			name: "nil",
		},
		{
			name: "string/interface map, no id",
			i: map[string]interface{}{
				"value": "foo",
			},
		},
		{
			name: "string/interface map, with id",
			i: map[string]interface{}{
				"_id": "foo",
			},
			id:       "foo",
			expected: true,
		},
		{
			name: "string/string map, with id",
			i: map[string]string{
				"_id": "foo",
			},
			id:       "foo",
			expected: true,
		},
		{
			name: "invalid JSON",
			i:    make(chan int),
		},
		{
			name: "valid JSON",
			i: struct {
				ID string `json:"_id"`
			}{ID: "oink"},
			id:       "oink",
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, ok := extractDocID(test.i)
			if ok != test.expected || test.id != id {
				t.Errorf("Expected %t/%s, got %t/%s", test.expected, test.id, ok, id)
			}
		})
	}
}

func TestRowScanDoc(t *testing.T) {
	tests := []struct {
		name     string
		row      *Row
		expected interface{}
		status   int
		err      string
	}{
		{
			name:   "invalid json",
			row:    &Row{doc: []byte("invalid json")},
			status: StatusBadResponse,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name:     "success",
			row:      &Row{doc: []byte(`{"foo":123.4}`)},
			expected: map[string]interface{}{"foo": 123.4},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := test.row.ScanDoc(&result)
			testy.StatusError(t, test.err, test.status, err)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestCreateDoc(t *testing.T) {
	tests := []struct {
		name       string
		db         *DB
		doc        interface{}
		options    Options
		docID, rev string
		status     int
		err        string
	}{
		{
			name: "error",
			db: &DB{
				driverDB: &mockDBOpts{
					CreateDocOptsFunc: func(_ context.Context, _ interface{}, _ map[string]interface{}) (string, string, error) {
						return "", "", errors.Status(StatusBadRequest, "create error")
					},
				},
			},
			status: StatusBadRequest,
			err:    "create error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mockDBOpts{
					CreateDocOptsFunc: func(_ context.Context, doc interface{}, opts map[string]interface{}) (string, string, error) {
						expectedDoc := map[string]string{"type": "test"}
						if d := diff.Interface(expectedDoc, doc); d != nil {
							return "", "", fmt.Errorf("Unexpected doc:\n%s", d)
						}
						if d := diff.Interface(testOptions, opts); d != nil {
							return "", "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "foo", "1-xxx", nil
					},
				},
			},
			doc:     map[string]string{"type": "test"},
			options: testOptions,
			docID:   "foo",
			rev:     "1-xxx",
		},
		{
			name: "legacy",
			db: &DB{
				driverDB: &mockDB{
					CreateDocFunc: func(_ context.Context, doc interface{}) (string, string, error) {
						expectedDoc := map[string]string{"type": "test"}
						if d := diff.Interface(expectedDoc, doc); d != nil {
							return "", "", fmt.Errorf("Unexpected doc:\n%s", d)
						}
						return "foo", "1-xxx", nil
					},
				},
			},
			doc:   map[string]string{"type": "test"},
			docID: "foo",
			rev:   "1-xxx",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			docID, rev, err := test.db.CreateDoc(context.Background(), test.doc, test.options)
			testy.StatusError(t, test.err, test.status, err)
			if docID != test.docID || test.rev != test.rev {
				t.Errorf("Unexpected result: %s / %s", docID, rev)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name       string
		db         *DB
		docID, rev string
		options    Options
		newRev     string
		status     int
		err        string
	}{
		{
			name:   "no doc ID",
			status: StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name: "error",
			db: &DB{
				driverDB: &mockDBOpts{
					DeleteOptsFunc: func(_ context.Context, _, _ string, _ map[string]interface{}) (string, error) {
						return "", errors.Status(StatusBadRequest, "delete error")
					},
				},
			},
			docID:  "foo",
			status: StatusBadRequest,
			err:    "delete error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mockDBOpts{
					DeleteOptsFunc: func(_ context.Context, docID, rev string, opts map[string]interface{}) (string, error) {
						expectedDocID := "foo"
						expectedRev := "1-xxx"
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						if rev != expectedRev {
							return "", fmt.Errorf("Unexpected rev: %s", rev)
						}
						if d := diff.Interface(testOptions, opts); d != nil {
							return "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "2-xxx", nil
					},
				},
			},
			docID:   "foo",
			rev:     "1-xxx",
			options: testOptions,
			newRev:  "2-xxx",
		},
		{
			name: "legacy",
			db: &DB{
				driverDB: &mockDB{
					DeleteFunc: func(_ context.Context, docID, rev string) (string, error) {
						expectedDocID := "foo"
						expectedRev := "1-xxx"
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						if rev != expectedRev {
							return "", fmt.Errorf("Unexpected rev: %s", rev)
						}
						return "2-xxx", nil
					},
				},
			},
			docID:  "foo",
			rev:    "1-xxx",
			newRev: "2-xxx",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			newRev, err := test.db.Delete(context.Background(), test.docID, test.rev, test.options)
			testy.StatusError(t, test.err, test.status, err)
			if newRev != test.newRev {
				t.Errorf("Unexpected newRev: %s", newRev)
			}
		})
	}
}

type legacyPutAttRecorder struct {
	driver.DB
	docID, rev, filename, cType string
	body                        string

	newRev string
	err    error
}

func (db *legacyPutAttRecorder) PutAttachment(_ context.Context, docID, rev, filename, contentType string, body io.Reader) (string, error) {
	if db.docID != docID || db.rev != rev {
		return "", errors.Errorf("Unexpected id/rev: %s/%s", docID, rev)
	}
	if db.filename != filename || db.cType != contentType {
		return "", errors.Errorf("Unexpected file data: %s / %s", filename, contentType)
	}
	content, err := ioutil.ReadAll(body)
	if err != nil {
		panic(err)
	}
	if d := diff.Text(db.body, string(content)); d != nil {
		return "", errors.Errorf("Unexpected content: %s", d)
	}
	return db.newRev, db.err
}

type putAttRecorder struct {
	driver.DB
	driver.DBOpts
	docID, rev, filename, cType string
	body                        string
	opts                        map[string]interface{}

	newRev string
	err    error
}

func (db *putAttRecorder) PutAttachment(_ context.Context, _, _, _, _ string, _ io.Reader) (string, error) {
	panic("PutAttachment called")
}

func (db *putAttRecorder) PutAttachmentOpts(_ context.Context, docID, rev, filename, contentType string, body io.Reader, opts map[string]interface{}) (string, error) {
	if db.docID != docID || db.rev != rev {
		return "", errors.Errorf("Unexpected id/rev: %s/%s", docID, rev)
	}
	if db.filename != filename || db.cType != contentType {
		return "", errors.Errorf("Unexpected file data: %s / %s", filename, contentType)
	}
	content, err := ioutil.ReadAll(body)
	if err != nil {
		panic(err)
	}
	if d := diff.Text(db.body, string(content)); d != nil {
		return "", errors.Errorf("Unexpected content: %s", d)
	}
	if d := diff.Interface(db.opts, opts); d != nil {
		return "", errors.Errorf("Unexpected options: %s", d)
	}

	return db.newRev, db.err
}

func TestPutAttachment(t *testing.T) {
	tests := []struct {
		name       string
		db         *DB
		docID, rev string
		att        *Attachment
		options    Options
		newRev     string
		status     int
		err        string

		body string
	}{
		{
			name:  "db error",
			docID: "foo",
			db: &DB{driverDB: &putAttRecorder{
				docID:    "foo",
				filename: "foo.txt",
				err:      errors.Status(StatusBadRequest, "db error"),
			}},
			att: &Attachment{
				Filename:   "foo.txt",
				ReadCloser: ioutil.NopCloser(strings.NewReader("")),
			},
			status: StatusBadRequest,
			err:    "db error",
		},
		{
			name:   "no doc id",
			status: StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name:   "no filename",
			docID:  "foo",
			att:    &Attachment{},
			status: StatusBadRequest,
			err:    "kivik: filename required",
		},
		{
			name:  "success",
			docID: "foo",
			rev:   "1-xxx",
			db: &DB{driverDB: &putAttRecorder{
				docID:    "foo",
				rev:      "1-xxx",
				filename: "foo.txt",
				cType:    "text/plain",
				body:     "Test file",
				opts:     map[string]interface{}{"opt": 1},
				newRev:   "2-xxx",
			}},
			att: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				ReadCloser:  ioutil.NopCloser(strings.NewReader("Test file")),
			},
			options: Options{"opt": 1},
			newRev:  "2-xxx",
			body:    "Test file",
		},
		{
			name:  "legacy",
			docID: "foo",
			rev:   "1-xxx",
			db: &DB{driverDB: &legacyPutAttRecorder{
				docID:    "foo",
				rev:      "1-xxx",
				filename: "foo.txt",
				cType:    "text/plain",
				body:     "Test file",
				newRev:   "2-xxx",
			}},
			att: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				ReadCloser:  ioutil.NopCloser(strings.NewReader("Test file")),
			},
			newRev: "2-xxx",
			body:   "Test file",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			newRev, err := test.db.PutAttachment(context.Background(), test.docID, test.rev, test.att, test.options)
			testy.StatusError(t, test.err, test.status, err)
			if newRev != test.newRev {
				t.Errorf("Unexpected newRev: %s", newRev)
			}
		})
	}
}

func TestDeleteAttachment(t *testing.T) {
	tests := []struct {
		name                 string
		db                   *DB
		docID, rev, filename string
		options              Options

		newRev string
		status int
		err    string
	}{
		{
			name:   "missing doc id",
			status: StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name:   "missing filename",
			docID:  "foo",
			status: StatusBadRequest,
			err:    "kivik: filename required",
		},
		{
			name:     "db error",
			docID:    "foo",
			filename: "foo.txt",
			db: &DB{
				driverDB: &mockDBOpts{
					DeleteAttachmentOptsFunc: func(_ context.Context, _, _, _ string, _ map[string]interface{}) (string, error) {
						return "", errors.Status(StatusBadRequest, "db error")
					},
				},
			},
			status: StatusBadRequest,
			err:    "db error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mockDBOpts{
					DeleteAttachmentOptsFunc: func(_ context.Context, docID, rev, filename string, opts map[string]interface{}) (string, error) {
						expectedDocID, expectedRev, expectedFilename := "foo", "1-xxx", "foo.txt"
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						if rev != expectedRev {
							return "", fmt.Errorf("Unexpected rev: %s", rev)
						}
						if filename != expectedFilename {
							return "", fmt.Errorf("Unexpected filename: %s", filename)
						}
						if d := diff.Interface(testOptions, opts); d != nil {
							return "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "2-xxx", nil
					},
				},
			},
			docID:    "foo",
			rev:      "1-xxx",
			filename: "foo.txt",
			options:  testOptions,
			newRev:   "2-xxx",
		},
		{
			name: "legacy",
			db: &DB{
				driverDB: &mockDB{
					DeleteAttachmentFunc: func(_ context.Context, docID, rev, filename string) (string, error) {
						expectedDocID, expectedRev, expectedFilename := "foo", "1-xxx", "foo.txt"
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						if rev != expectedRev {
							return "", fmt.Errorf("Unexpected rev: %s", rev)
						}
						if filename != expectedFilename {
							return "", fmt.Errorf("Unexpected filename: %s", filename)
						}
						return "2-xxx", nil
					},
				},
			},
			docID:    "foo",
			rev:      "1-xxx",
			filename: "foo.txt",
			newRev:   "2-xxx",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			newRev, err := test.db.DeleteAttachment(context.Background(), test.docID, test.rev, test.filename, test.options)
			testy.StatusError(t, test.err, test.status, err)
			if newRev != test.newRev {
				t.Errorf("Unexpected new rev: %s", newRev)
			}
		})
	}
}

type mockOldAttGetter struct {
	driver.DB
	docID, rev, filename string

	cType   string
	md5     driver.MD5sum
	content io.ReadCloser
	err     error
}

func (db *mockOldAttGetter) GetAttachment(_ context.Context, docID, rev, filename string) (string, driver.MD5sum, io.ReadCloser, error) {
	if docID != db.docID {
		return "", driver.MD5sum{}, nil, errors.Errorf("Unexpected docID: %s", docID)
	}
	if rev != db.rev {
		return "", driver.MD5sum{}, nil, errors.Errorf("Unexpected rev: %s", rev)
	}
	if filename != db.filename {
		return "", driver.MD5sum{}, nil, errors.Errorf("Unexpected filename: %s", filename)
	}
	return db.cType, db.md5, db.content, db.err
}

type mockAttGetter struct {
	driver.DB
	driver.DBOpts
	docID, rev, filename string
	opts                 map[string]interface{}

	cType   string
	md5     driver.MD5sum
	content io.ReadCloser
	err     error
}

func (db *mockAttGetter) GetAttachmentOpts(_ context.Context, docID, rev, filename string, opts map[string]interface{}) (string, driver.MD5sum, io.ReadCloser, error) {
	if docID != db.docID {
		return "", driver.MD5sum{}, nil, errors.Errorf("Unexpected docID: %s", docID)
	}
	if rev != db.rev {
		return "", driver.MD5sum{}, nil, errors.Errorf("Unexpected rev: %s", rev)
	}
	if filename != db.filename {
		return "", driver.MD5sum{}, nil, errors.Errorf("Unexpected filename: %s", filename)
	}
	if d := diff.Interface(db.opts, opts); d != nil {
		return "", driver.MD5sum{}, nil, errors.Errorf("Unexpected options: %s", d)
	}
	return db.cType, db.md5, db.content, db.err
}

func TestGetAttachment(t *testing.T) {
	tests := []struct {
		name                 string
		db                   *DB
		docID, rev, filename string
		options              Options

		content  string
		expected *Attachment
		status   int
		err      string
	}{
		{
			name: "legacy, error",
			db: &DB{driverDB: &mockOldAttGetter{
				docID:    "foo",
				filename: "foo.txt",
				err:      errors.New("fail"),
			}},
			docID:    "foo",
			filename: "foo.txt",
			status:   500,
			err:      "fail",
		},
		{
			name: "legacy, success",
			db: &DB{driverDB: &mockOldAttGetter{
				docID:    "foo",
				rev:      "1-xxx",
				filename: "foo.txt",
				cType:    "text/plain",
				md5:      driver.MD5sum{0x01},
				content:  body("Test"),
			}},
			docID:    "foo",
			rev:      "1-xxx",
			filename: "foo.txt",
			content:  "Test",
			expected: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				MD5:         driver.MD5sum{0x01},
			},
		},
		{
			name: "new, error",
			db: &DB{driverDB: &mockAttGetter{
				docID:    "foo",
				filename: "foo.txt",
				err:      errors.New("fail"),
			}},
			docID:    "foo",
			filename: "foo.txt",
			status:   500,
			err:      "fail",
		},
		{
			name: "new, success",
			db: &DB{driverDB: &mockAttGetter{
				docID:    "foo",
				rev:      "1-xxx",
				filename: "foo.txt",
				opts:     map[string]interface{}{"foo": "bar"},
				cType:    "text/plain",
				md5:      driver.MD5sum{0x01},
				content:  body("Test"),
			}},
			docID:    "foo",
			rev:      "1-xxx",
			filename: "foo.txt",
			options:  Options{"foo": "bar"},
			content:  "Test",
			expected: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				MD5:         driver.MD5sum{0x01},
			},
		},
		{
			name:   "no docID",
			status: StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name:   "no filename",
			docID:  "foo",
			status: StatusBadRequest,
			err:    "kivik: filename required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.GetAttachment(context.Background(), test.docID, test.rev, test.filename, test.options)
			testy.StatusError(t, test.err, test.status, err)
			content, err := ioutil.ReadAll(result)
			if err != nil {
				t.Fatal(err)
			}
			if d := diff.Text(test.content, string(content)); d != nil {
				t.Errorf("Unexpected content:\n%s", d)
			}
			result.ReadCloser = nil
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

type mockOldAttMetaer struct {
	driver.DB
	docID, rev, filename string

	cType string
	md5   driver.MD5sum
	err   error
}

var _ driver.OldAttachmentMetaer = &mockOldAttMetaer{}

func (db *mockOldAttMetaer) GetAttachmentMeta(_ context.Context, docID, rev, filename string) (string, driver.MD5sum, error) {
	if docID != db.docID {
		return "", driver.MD5sum{}, errors.Errorf("Unexpected docID: %s", docID)
	}
	if rev != db.rev {
		return "", driver.MD5sum{}, errors.Errorf("Unexpected rev: %s", rev)
	}
	if filename != db.filename {
		return "", driver.MD5sum{}, errors.Errorf("Unexpected filename: %s", filename)
	}
	return db.cType, db.md5, db.err
}

type mockAttMetaer struct {
	driver.DB
	docID, rev, filename string
	opts                 map[string]interface{}

	cType string
	md5   driver.MD5sum
	err   error
}

var _ driver.AttachmentMetaer = &mockAttMetaer{}

func (db *mockAttMetaer) GetAttachmentMeta(_ context.Context, docID, rev, filename string, opts map[string]interface{}) (string, driver.MD5sum, error) {
	if docID != db.docID {
		return "", driver.MD5sum{}, errors.Errorf("Unexpected docID: %s", docID)
	}
	if rev != db.rev {
		return "", driver.MD5sum{}, errors.Errorf("Unexpected rev: %s", rev)
	}
	if filename != db.filename {
		return "", driver.MD5sum{}, errors.Errorf("Unexpected filename: %s", filename)
	}
	if d := diff.Interface(db.opts, opts); d != nil {
		return "", driver.MD5sum{}, errors.Errorf("Unexpected options: %s", d)
	}
	return db.cType, db.md5, db.err
}

func TestGetAttachmentMeta(t *testing.T) {
	tests := []struct {
		name                 string
		db                   *DB
		docID, rev, filename string
		options              Options

		expected *Attachment
		status   int
		err      string
	}{
		{
			name: "plain db, error",
			db: &DB{driverDB: &mockAttGetter{
				docID:    "foo",
				filename: "foo.txt",
				err:      errors.New("fail"),
			}},
			docID:    "foo",
			filename: "foo.txt",
			status:   500,
			err:      "fail",
		},
		{
			name: "plain db, success",
			db: &DB{driverDB: &mockAttGetter{
				docID:    "foo",
				rev:      "1-xxx",
				filename: "foo.txt",
				cType:    "text/plain",
				md5:      driver.MD5sum{0x01},
				content:  body("Test"),
			}},
			docID:    "foo",
			rev:      "1-xxx",
			filename: "foo.txt",
			expected: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				MD5:         driver.MD5sum{0x01},
			},
		},
		{
			name: "legacy metaer, error",
			db: &DB{driverDB: &mockOldAttMetaer{
				docID:    "foo",
				filename: "foo.txt",
				err:      errors.New("fail"),
			}},
			docID:    "foo",
			filename: "foo.txt",
			status:   500,
			err:      "fail",
		},
		{
			name: "legacy metaer, success",
			db: &DB{driverDB: &mockOldAttMetaer{
				docID:    "foo",
				rev:      "1-xxx",
				filename: "foo.txt",
				cType:    "text/plain",
				md5:      driver.MD5sum{0x01},
			}},
			docID:    "foo",
			rev:      "1-xxx",
			filename: "foo.txt",
			expected: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				MD5:         driver.MD5sum{0x01},
			},
		},
		{
			name: "new metaer, error",
			db: &DB{driverDB: &mockAttMetaer{
				docID:    "foo",
				filename: "foo.txt",
				err:      errors.New("fail"),
			}},
			docID:    "foo",
			filename: "foo.txt",
			status:   500,
			err:      "fail",
		},
		{
			name: "new metaer, success",
			db: &DB{driverDB: &mockAttMetaer{
				docID:    "foo",
				rev:      "1-xxx",
				filename: "foo.txt",
				cType:    "text/plain",
				md5:      driver.MD5sum{0x01},
				opts:     map[string]interface{}{"foo": "bar"},
			}},
			docID:    "foo",
			rev:      "1-xxx",
			filename: "foo.txt",
			options:  Options{"foo": "bar"},
			expected: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				MD5:         driver.MD5sum{0x01},
			},
		},
		{
			name:   "no doc id",
			status: StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name:   "no filename",
			docID:  "foo",
			status: StatusBadRequest,
			err:    "kivik: filename required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.GetAttachmentMeta(context.Background(), test.docID, test.rev, test.filename, test.options)
			testy.StatusError(t, test.err, test.status, err)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
