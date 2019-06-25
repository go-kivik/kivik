package kivik

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"

	"github.com/go-kivik/kivik/driver"
	"github.com/go-kivik/kivik/mock"
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
				driverDB: &mock.DB{
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
				driverDB: &mock.DB{
					AllDocsFunc: func(_ context.Context, opts map[string]interface{}) (driver.Rows, error) {
						if d := diff.Interface(testOptions, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options: %s", d)
						}
						return &mock.Rows{ID: "a"}, nil
					},
				},
			},
			options: testOptions,
			expected: &Rows{
				iter: &iter{
					feed: &rowsIterator{
						Rows: &mock.Rows{ID: "a"},
					},
					curVal: &driver.Row{},
				},
				rowsi: &mock.Rows{ID: "a"},
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

func TestDesignDocs(t *testing.T) {
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
				driverDB: &mock.DesignDocer{
					DesignDocsFunc: func(_ context.Context, _ map[string]interface{}) (driver.Rows, error) {
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
				driverDB: &mock.DesignDocer{
					DesignDocsFunc: func(_ context.Context, opts map[string]interface{}) (driver.Rows, error) {
						if d := diff.Interface(testOptions, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options: %s", d)
						}
						return &mock.Rows{ID: "a"}, nil
					},
				},
			},
			options: testOptions,
			expected: &Rows{
				iter: &iter{
					feed: &rowsIterator{
						Rows: &mock.Rows{ID: "a"},
					},
					curVal: &driver.Row{},
				},
				rowsi: &mock.Rows{ID: "a"},
			},
		},
		{
			name:   "not supported",
			db:     &DB{driverDB: &mock.DB{}},
			status: StatusNotImplemented,
			err:    "kivik: design doc view not supported by driver",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.DesignDocs(context.Background(), test.options)
			testy.StatusError(t, test.err, test.status, err)
			result.cancel = nil // Determinism
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
func TestLocalDocs(t *testing.T) {
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
				driverDB: &mock.LocalDocer{
					LocalDocsFunc: func(_ context.Context, _ map[string]interface{}) (driver.Rows, error) {
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
				driverDB: &mock.LocalDocer{
					LocalDocsFunc: func(_ context.Context, opts map[string]interface{}) (driver.Rows, error) {
						if d := diff.Interface(testOptions, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options: %s", d)
						}
						return &mock.Rows{ID: "a"}, nil
					},
				},
			},
			options: testOptions,
			expected: &Rows{
				iter: &iter{
					feed: &rowsIterator{
						Rows: &mock.Rows{ID: "a"},
					},
					curVal: &driver.Row{},
				},
				rowsi: &mock.Rows{ID: "a"},
			},
		},
		{
			name:   "not supported",
			db:     &DB{driverDB: &mock.DB{}},
			status: StatusNotImplemented,
			err:    "kivik: local doc view not supported by driver",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.LocalDocs(context.Background(), test.options)
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
				driverDB: &mock.DB{
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
				driverDB: &mock.DB{
					QueryFunc: func(_ context.Context, ddoc, view string, opts map[string]interface{}) (driver.Rows, error) {
						expectedDdoc := "foo"
						expectedView := "bar" // nolint: goconst
						if ddoc != expectedDdoc {
							return nil, fmt.Errorf("Unexpected ddoc: %s", ddoc)
						}
						if view != expectedView {
							return nil, fmt.Errorf("Unexpected view: %s", view)
						}
						if d := diff.Interface(testOptions, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options: %s", d)
						}
						return &mock.Rows{ID: "a"}, nil
					},
				},
			},
			ddoc:    "foo",
			view:    "bar",
			options: testOptions,
			expected: &Rows{
				iter: &iter{
					feed: &rowsIterator{
						Rows: &mock.Rows{ID: "a"},
					},
					curVal: &driver.Row{},
				},
				rowsi: &mock.Rows{ID: "a"},
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
	}{
		{
			name: "db error",
			db: &DB{
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, _ string, _ map[string]interface{}) (*driver.Document, error) {
						return nil, fmt.Errorf("db error")
					},
				},
			},
			expected: &Row{
				Err: fmt.Errorf("db error"),
			},
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, docID string, options map[string]interface{}) (*driver.Document, error) {
						expectedDocID := "foo"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						if d := diff.Interface(testOptions, options); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%s", d)
						}
						return &driver.Document{
							ContentLength: 13,
							Rev:           "1-xxx",
							Body:          body(`{"_id":"foo"}`),
						}, nil
					},
				},
			},
			docID:   "foo",
			options: testOptions,
			expected: &Row{
				ContentLength: 13,
				Rev:           "1-xxx",
				Body:          body(`{"_id":"foo"}`),
			},
		},
		{
			name: "streaming attachments",
			db: &DB{
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, docID string, options map[string]interface{}) (*driver.Document, error) {
						expectedDocID := "foo"
						expectedOptions := map[string]interface{}{"include_docs": true}
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						if d := diff.Interface(expectedOptions, options); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%s", d)
						}
						return &driver.Document{
							ContentLength: 13,
							Rev:           "1-xxx",
							Body:          body(`{"_id":"foo"}`),
							Attachments:   &mock.Attachments{ID: "asdf"},
						}, nil
					},
				},
			},
			docID:   "foo",
			options: map[string]interface{}{"include_docs": true},
			expected: &Row{
				ContentLength: 13,
				Rev:           "1-xxx",
				Body:          body(`{"_id":"foo"}`),
				Attachments: &AttachmentsIterator{
					atti: &mock.Attachments{ID: "asdf"},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.db.Get(context.Background(), test.docID, test.options)
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
				driverDB: &mock.DB{},
			},
			status: StatusNotImplemented,
			err:    "kivik: flush not supported by driver",
		},
		{
			name: "db error",
			db: &DB{
				driverDB: &mock.Flusher{
					FlushFunc: func(_ context.Context) error {
						return &Error{HTTPStatus: http.StatusBadGateway, Err: errors.New("flush error")}
					},
				},
			},
			status: StatusBadResponse,
			err:    "flush error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mock.Flusher{
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
				driverDB: &mock.DB{
					StatsFunc: func(_ context.Context) (*driver.DBStats, error) {
						return nil, &Error{HTTPStatus: http.StatusBadGateway, Err: errors.New("stats error")}
					},
				},
			},
			status: StatusBadResponse,
			err:    "stats error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mock.DB{
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
		driverDB: &mock.DB{
			CompactFunc: func(_ context.Context) error {
				return &Error{HTTPStatus: http.StatusBadRequest, Err: errors.New(expected)}
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
		driverDB: &mock.DB{
			CompactViewFunc: func(_ context.Context, ddocID string) error {
				if ddocID != expectedDDocID {
					return fmt.Errorf("Unexpected ddocID: %s", ddocID)
				}
				return &Error{HTTPStatus: http.StatusBadRequest, Err: errors.New(expected)}
			},
		},
	}
	err := db.CompactView(context.Background(), expectedDDocID)
	testy.StatusError(t, expected, StatusBadRequest, err)
}

func TestViewCleanup(t *testing.T) {
	expected := "compact error"
	db := &DB{
		driverDB: &mock.DB{
			ViewCleanupFunc: func(_ context.Context) error {
				return &Error{HTTPStatus: http.StatusBadRequest, Err: errors.New(expected)}
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
				driverDB: &mock.DB{
					SecurityFunc: func(_ context.Context) (*driver.Security, error) {
						return nil, &Error{HTTPStatus: http.StatusBadGateway, Err: errors.New("security error")}
					},
				},
			},
			status: StatusBadResponse,
			err:    "security error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mock.DB{
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
			db:     &DB{},
			status: StatusBadRequest,
			err:    "kivik: security required",
		},
		{
			name: "set error",
			db: &DB{
				driverDB: &mock.DB{
					SetSecurityFunc: func(_ context.Context, _ *driver.Security) error {
						return &Error{HTTPStatus: http.StatusBadGateway, Err: errors.New("set security error")}
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
				driverDB: &mock.DB{
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

func TestGetMeta(t *testing.T) { // nolint: gocyclo
	tests := []struct {
		name    string
		db      *DB
		docID   string
		size    int64
		rev     string
		options Options
		status  int
		err     string
	}{
		{
			name: "meta getter error",
			db: &DB{
				driverDB: &mock.MetaGetter{
					GetMetaFunc: func(_ context.Context, _ string, _ map[string]interface{}) (int64, string, error) {
						return 0, "", &Error{HTTPStatus: http.StatusBadGateway, Err: errors.New("get meta error")}
					},
				},
			},
			status: StatusBadResponse,
			err:    "get meta error",
		},
		{
			name: "meta getter success",
			db: &DB{
				driverDB: &mock.MetaGetter{
					GetMetaFunc: func(_ context.Context, docID string, opts map[string]interface{}) (int64, string, error) {
						expectedDocID := "foo"
						if docID != expectedDocID {
							return 0, "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						if d := diff.Interface(testOptions, opts); d != nil {
							return 0, "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return 123, "1-xxx", nil
					},
				},
			},
			docID:   "foo",
			options: testOptions,
			size:    123,
			rev:     "1-xxx",
		},
		{
			name: "non-meta getter error",
			db: &DB{
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, _ string, _ map[string]interface{}) (*driver.Document, error) {
						return nil, &Error{HTTPStatus: http.StatusBadGateway, Err: errors.New("get error")}
					},
				},
			},
			status: StatusBadResponse,
			err:    "get error",
		},
		{
			name: "non-meta getter success with rev",
			db: &DB{
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, docID string, opts map[string]interface{}) (*driver.Document, error) {
						expectedDocID := "foo"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						if opts != nil {
							return nil, errors.New("opts should be nil")
						}
						return &driver.Document{
							ContentLength: 16,
							Rev:           "1-xxx",
							Body:          body(`{"_rev":"1-xxx"}`),
						}, nil
					},
				},
			},
			docID: "foo",
			size:  16,
			rev:   "1-xxx",
		},
		{
			name: "non-meta getter success without rev",
			db: &DB{
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, docID string, opts map[string]interface{}) (*driver.Document, error) {
						expectedDocID := "foo"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						if opts != nil {
							return nil, errors.New("opts should be nil")
						}
						return &driver.Document{
							ContentLength: 16,
							Body:          body(`{"_rev":"1-xxx"}`),
						}, nil
					},
				},
			},
			docID: "foo",
			size:  16,
			rev:   "1-xxx",
		},
		{
			name: "non-meta getter success without rev, invalid json",
			db: &DB{
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, docID string, opts map[string]interface{}) (*driver.Document, error) {
						expectedDocID := "foo"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						if opts != nil {
							return nil, errors.New("opts should be nil")
						}
						return &driver.Document{
							ContentLength: 16,
							Body:          body(`invalid json`),
						}, nil
					},
				},
			},
			docID:  "foo",
			status: http.StatusInternalServerError,
			err:    "invalid character 'i' looking for beginning of value",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			size, rev, err := test.db.GetMeta(context.Background(), test.docID, test.options)
			testy.StatusError(t, test.err, test.status, err)
			if size != test.size {
				t.Errorf("Unexpected size: %v", size)
			}
			if rev != test.rev {
				t.Errorf("Unexpected rev: %v", rev)
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
			db:     &DB{},
			status: StatusBadRequest,
			err:    "kivik: targetID required",
		},
		{
			name:   "missing source",
			db:     &DB{},
			target: "foo",
			status: StatusBadRequest,
			err:    "kivik: sourceID required",
		},
		{
			name: "copier error",
			db: &DB{
				driverDB: &mock.Copier{
					CopyFunc: func(_ context.Context, _, _ string, _ map[string]interface{}) (string, error) {
						return "", &Error{HTTPStatus: http.StatusBadRequest, Err: errors.New("copy error")}
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
				driverDB: &mock.Copier{
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
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, _ string, _ map[string]interface{}) (*driver.Document, error) {
						return nil, &Error{HTTPStatus: http.StatusBadGateway, Err: errors.New("get error")}
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
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, _ string, _ map[string]interface{}) (*driver.Document, error) {
						return &driver.Document{
							ContentLength: 12,
							Body:          body("invalid json"),
						}, nil
					},
				},
			},
			target: "foo",
			source: "bar",
			status: http.StatusInternalServerError,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name: "non-copier put error",
			db: &DB{
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, _ string, _ map[string]interface{}) (*driver.Document, error) {
						return &driver.Document{
							ContentLength: 28,
							Body:          body(`{"_id":"foo","_rev":"1-xxx"}`),
						}, nil
					},
					PutFunc: func(_ context.Context, _ string, _ interface{}, _ map[string]interface{}) (string, error) {
						return "", &Error{HTTPStatus: http.StatusBadGateway, Err: errors.New("put error")}
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
				driverDB: &mock.DB{
					GetFunc: func(_ context.Context, docID string, options map[string]interface{}) (*driver.Document, error) {
						expectedDocID := "bar"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected get docID: %s", docID)
						}
						return &driver.Document{
							ContentLength: 40,
							Body:          body(`{"_id":"bar","_rev":"1-xxx","foo":123.4}`),
						}, nil
					},
					PutFunc: func(_ context.Context, docID string, doc interface{}, opts map[string]interface{}) (string, error) {
						expectedDocID := "foo"
						expectedDoc := map[string]interface{}{"_id": "foo", "foo": 123.4}
						expectedOpts := map[string]interface{}{"batch": true}
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected put docID: %s", docID)
						}
						if d := diff.Interface(expectedDoc, doc); d != nil {
							return "", fmt.Errorf("Unexpected doc:\n%s", doc)
						}
						if d := diff.Interface(expectedOpts, opts); d != nil {
							return "", fmt.Errorf("Unexpected opts:\n%s", opts)
						}
						return "1-xxx", nil
					},
				},
			},
			target:   "foo",
			source:   "bar",
			options:  Options{"rev": "1-xxx", "batch": true},
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
			Status: StatusBadAPICall,
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
			Status: http.StatusBadRequest,
			Error:  "errorReader",
		},
	}
	for _, test := range tests {
		func(test njTest) {
			t.Run(test.Name, func(t *testing.T) {
				result, err := normalizeFromJSON(test.Input)
				testy.StatusError(t, test.Error, test.Status, err)
				if d := diff.Interface(test.Expected, result); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}

func TestPut(t *testing.T) {
	putFunc := func(_ context.Context, docID string, doc interface{}, opts map[string]interface{}) (string, error) {
		expectedDocID := "foo"
		expectedDoc := map[string]interface{}{"foo": "bar"}
		if expectedDocID != docID {
			return "", fmt.Errorf("Unexpected docID: %s", docID)
		}
		if d := diff.Interface(expectedDoc, doc); d != nil {
			return "", fmt.Errorf("Unexpected doc: %s", d)
		}
		if d := diff.Interface(testOptions, opts); d != nil {
			return "", fmt.Errorf("Unexpected opts: %s", d)
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
			db:     &DB{},
			status: StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name: "db error",
			db: &DB{
				driverDB: &mock.DB{
					PutFunc: func(_ context.Context, _ string, _ interface{}, _ map[string]interface{}) (string, error) {
						return "", &Error{HTTPStatus: http.StatusBadRequest, Err: errors.New("db error")}
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
				driverDB: &mock.DB{
					PutFunc: putFunc,
				},
			},
			docID:   "foo",
			input:   map[string]interface{}{"foo": "bar"},
			options: testOptions,
			newRev:  "1-xxx",
		},
		{
			name:   "InvalidJSON",
			db:     &DB{},
			docID:  "foo",
			input:  []byte("Something bogus"),
			status: StatusBadAPICall,
			err:    "invalid character 'S' looking for beginning of value",
		},
		{
			name: "Bytes",
			db: &DB{
				driverDB: &mock.DB{
					PutFunc: putFunc,
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
				driverDB: &mock.DB{
					PutFunc: putFunc,
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
				driverDB: &mock.DB{
					PutFunc: putFunc,
				},
			},
			docID:   "foo",
			input:   strings.NewReader(`{"foo":"bar"}`),
			options: testOptions,
			newRev:  "1-xxx",
		},
		{
			name:   "ErrorReader",
			db:     &DB{},
			docID:  "foo",
			input:  &errorReader{},
			status: http.StatusBadRequest,
			err:    "errorReader",
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
		dst      interface{}
		expected interface{}
		status   int
		err      string
	}{
		{
			name:   "non-pointer dst",
			row:    &Row{Body: body(`{"foo":123.4}`)},
			dst:    map[string]interface{}{},
			status: http.StatusBadRequest,
			err:    "kivik: destination is not a pointer",
		},
		{
			name:   "invalid json",
			row:    &Row{Body: body("invalid json")},
			dst:    new(map[string]interface{}),
			status: http.StatusInternalServerError,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name:     "success",
			row:      &Row{Body: body(`{"foo":123.4}`)},
			dst:      new(map[string]interface{}),
			expected: &map[string]interface{}{"foo": 123.4},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.row.ScanDoc(test.dst)
			testy.StatusError(t, test.err, test.status, err)
			if d := diff.Interface(test.expected, test.dst); d != nil {
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
				driverDB: &mock.DB{
					CreateDocFunc: func(_ context.Context, _ interface{}, _ map[string]interface{}) (string, string, error) {
						return "", "", &Error{HTTPStatus: http.StatusBadRequest, Err: errors.New("create error")}
					},
				},
			},
			status: StatusBadRequest,
			err:    "create error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mock.DB{
					CreateDocFunc: func(_ context.Context, doc interface{}, opts map[string]interface{}) (string, string, error) {
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			docID, rev, err := test.db.CreateDoc(context.Background(), test.doc, test.options)
			testy.StatusError(t, test.err, test.status, err)
			if docID != test.docID || rev != test.rev {
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
			db:     &DB{},
			status: StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name: "error",
			db: &DB{
				driverDB: &mock.DB{
					DeleteFunc: func(_ context.Context, _, _ string, _ map[string]interface{}) (string, error) {
						return "", &Error{HTTPStatus: http.StatusBadRequest, Err: errors.New("delete error")}
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
				driverDB: &mock.DB{
					DeleteFunc: func(_ context.Context, docID, rev string, opts map[string]interface{}) (string, error) {
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
			db: &DB{
				driverDB: &mock.DB{
					PutAttachmentFunc: func(_ context.Context, _, _ string, _ *driver.Attachment, _ map[string]interface{}) (string, error) {
						return "", &Error{HTTPStatus: http.StatusBadRequest, Err: errors.New("db error")}
					},
				},
			},
			att: &Attachment{
				Filename: "foo.txt",
				Content:  ioutil.NopCloser(strings.NewReader("")),
			},
			status: StatusBadRequest,
			err:    "db error",
		},
		{
			name:   "no doc id",
			db:     &DB{},
			status: StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name:   "no filename",
			db:     &DB{},
			docID:  "foo",
			att:    &Attachment{},
			status: StatusBadRequest,
			err:    "kivik: filename required",
		},
		{
			name:  "success",
			docID: "foo",
			rev:   "1-xxx",
			db: &DB{
				driverDB: &mock.DB{
					PutAttachmentFunc: func(_ context.Context, docID, rev string, att *driver.Attachment, opts map[string]interface{}) (string, error) {
						expectedDocID, expectedRev := "foo", "1-xxx"
						expectedContent := "Test file"
						expectedAtt := &driver.Attachment{
							Filename:    "foo.txt",
							ContentType: "text/plain",
						}
						if docID != expectedDocID {
							return "", fmt.Errorf("Unexpected docID: %s", docID)
						}
						if rev != expectedRev {
							return "", fmt.Errorf("Unexpected rev: %s", rev)
						}
						content, err := ioutil.ReadAll(att.Content)
						if err != nil {
							t.Fatal(err)
						}
						if d := diff.Text(expectedContent, string(content)); d != nil {
							return "", fmt.Errorf("Unexpected content:\n%s", string(content))
						}
						att.Content = nil
						if d := diff.Interface(expectedAtt, att); d != nil {
							return "", fmt.Errorf("Unexpected attachment:\n%s", d)
						}
						if d := diff.Interface(testOptions, opts); d != nil {
							return "", fmt.Errorf("Unexpected options:\n%s", d)
						}
						return "2-xxx", nil
					},
				},
			},
			att: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Content:     ioutil.NopCloser(strings.NewReader("Test file")),
			},
			options: testOptions,
			newRev:  "2-xxx",
			body:    "Test file",
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
			db:     &DB{},
			status: StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name:   "missing filename",
			db:     &DB{},
			docID:  "foo",
			status: StatusBadRequest,
			err:    "kivik: filename required",
		},
		{
			name:     "db error",
			docID:    "foo",
			filename: "foo.txt",
			db: &DB{
				driverDB: &mock.DB{
					DeleteAttachmentFunc: func(_ context.Context, _, _, _ string, _ map[string]interface{}) (string, error) {
						return "", &Error{HTTPStatus: http.StatusBadRequest, Err: errors.New("db error")}
					},
				},
			},
			status: StatusBadRequest,
			err:    "db error",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mock.DB{
					DeleteAttachmentFunc: func(_ context.Context, docID, rev, filename string, opts map[string]interface{}) (string, error) {
						expectedDocID, expectedRev, expectedFilename := "foo", "1-xxx", "foo.txt" // nolint: goconst
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

func TestGetAttachment(t *testing.T) {
	tests := []struct {
		name            string
		db              *DB
		docID, filename string
		options         Options

		content  string
		expected *Attachment
		status   int
		err      string
	}{
		{
			name: "error",
			db: &DB{
				driverDB: &mock.DB{
					GetAttachmentFunc: func(_ context.Context, _, _ string, _ map[string]interface{}) (*driver.Attachment, error) {
						return nil, errors.New("fail")
					},
				},
			},
			docID:    "foo",
			filename: "foo.txt",
			status:   500,
			err:      "fail",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mock.DB{
					GetAttachmentFunc: func(_ context.Context, docID, filename string, opts map[string]interface{}) (*driver.Attachment, error) {
						expectedDocID, expectedFilename := "foo", "foo.txt"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						if filename != expectedFilename {
							return nil, fmt.Errorf("Unexpected filename: %s", filename)
						}
						if d := diff.Interface(testOptions, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%s", d)
						}
						return &driver.Attachment{
							Filename:    "foo.txt",
							ContentType: "text/plain",
							Digest:      "md5-foo",
							Size:        4,
							Content:     body("Test"),
						}, nil
					},
				},
			},
			docID:    "foo",
			filename: "foo.txt",
			options:  testOptions,
			content:  "Test",
			expected: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Size:        4,
				Digest:      "md5-foo",
			},
		},
		{
			name:   "no docID",
			db:     &DB{},
			status: StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name:   "no filename",
			db:     &DB{},
			docID:  "foo",
			status: StatusBadRequest,
			err:    "kivik: filename required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.GetAttachment(context.Background(), test.docID, test.filename, test.options)
			testy.StatusError(t, test.err, test.status, err)
			content, err := ioutil.ReadAll(result.Content)
			if err != nil {
				t.Fatal(err)
			}
			if d := diff.Text(test.content, string(content)); d != nil {
				t.Errorf("Unexpected content:\n%s", d)
			}
			_ = result.Content.Close()
			result.Content = nil
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestGetAttachmentMeta(t *testing.T) { // nolint: gocyclo
	tests := []struct {
		name            string
		db              *DB
		docID, filename string
		options         Options

		expected *Attachment
		status   int
		err      string
	}{
		{
			name: "plain db, error",
			db: &DB{
				driverDB: &mock.DB{
					GetAttachmentFunc: func(_ context.Context, _, _ string, _ map[string]interface{}) (*driver.Attachment, error) {
						return nil, errors.New("fail")
					},
				},
			},
			docID:    "foo",
			filename: "foo.txt",
			status:   500,
			err:      "fail",
		},
		{
			name: "plain db, success",
			db: &DB{
				driverDB: &mock.DB{
					GetAttachmentFunc: func(_ context.Context, docID, filename string, opts map[string]interface{}) (*driver.Attachment, error) {
						expectedDocID, expectedFilename := "foo", "foo.txt"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						if filename != expectedFilename {
							return nil, fmt.Errorf("Unexpected filename: %s", filename)
						}
						if d := diff.Interface(testOptions, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%s", d)
						}
						return &driver.Attachment{
							Filename:    "foo.txt",
							ContentType: "text/plain",
							Digest:      "md5-foo",
							Size:        4,
							Content:     body("Test"),
						}, nil
					},
				},
			},
			docID:    "foo",
			filename: "foo.txt",
			options:  testOptions,
			expected: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Digest:      "md5-foo",
				Size:        4,
				Content:     nilContent,
			},
		},
		{
			name: "error",
			db: &DB{
				driverDB: &mock.AttachmentMetaGetter{
					GetAttachmentMetaFunc: func(_ context.Context, _, _ string, _ map[string]interface{}) (*driver.Attachment, error) {
						return nil, errors.New("fail")
					},
				},
			},
			docID:    "foo",
			filename: "foo.txt",
			status:   500,
			err:      "fail",
		},
		{
			name: "success",
			db: &DB{
				driverDB: &mock.AttachmentMetaGetter{
					GetAttachmentMetaFunc: func(_ context.Context, docID, filename string, opts map[string]interface{}) (*driver.Attachment, error) {
						expectedDocID, expectedFilename := "foo", "foo.txt"
						if docID != expectedDocID {
							return nil, fmt.Errorf("Unexpected docID: %s", docID)
						}
						if filename != expectedFilename {
							return nil, fmt.Errorf("Unexpected filename: %s", filename)
						}
						if d := diff.Interface(testOptions, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%s", d)
						}
						return &driver.Attachment{
							Filename:    "foo.txt",
							ContentType: "text/plain",
							Digest:      "md5-foo",
							Size:        4,
						}, nil
					},
				},
			},
			docID:    "foo",
			filename: "foo.txt",
			options:  testOptions,
			expected: &Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Digest:      "md5-foo",
				Size:        4,
				Content:     nilContent,
			},
		},
		{
			name:   "no doc id",
			db:     &DB{},
			status: StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name:   "no filename",
			db:     &DB{},
			docID:  "foo",
			status: StatusBadRequest,
			err:    "kivik: filename required",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.GetAttachmentMeta(context.Background(), test.docID, test.filename, test.options)
			testy.StatusError(t, test.err, test.status, err)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestPurge(t *testing.T) {
	type purgeTest struct {
		name   string
		db     *DB
		docMap map[string][]string

		expected *PurgeResult
		status   int
		err      string
	}

	docMap := map[string][]string{
		"foo": {"1-abc", "2-xyz"},
	}

	tests := []purgeTest{
		{
			name: "success, nothing purged",
			db: &DB{
				driverDB: &mock.Purger{
					PurgeFunc: func(_ context.Context, dm map[string][]string) (*driver.PurgeResult, error) {
						if d := diff.Interface(docMap, dm); d != nil {
							return nil, fmt.Errorf("Unexpected docmap: %s", d)
						}
						return &driver.PurgeResult{Seq: 2}, nil
					},
				},
			},
			docMap: docMap,
			expected: &PurgeResult{
				Seq: 2,
			},
		},
		{
			name: "success, all purged",
			db: &DB{
				driverDB: &mock.Purger{
					PurgeFunc: func(_ context.Context, dm map[string][]string) (*driver.PurgeResult, error) {
						if d := diff.Interface(docMap, dm); d != nil {
							return nil, fmt.Errorf("Unexpected docmap: %s", d)
						}
						return &driver.PurgeResult{Seq: 2, Purged: docMap}, nil
					},
				},
			},
			docMap: docMap,
			expected: &PurgeResult{
				Seq:    2,
				Purged: docMap,
			},
		},
		{
			name:   "non-purger",
			db:     &DB{driverDB: &mock.DB{}},
			status: StatusNotImplemented,
			err:    "kivik: purge not supported by driver",
		},
		{
			name: "couch 2.0-2.1 example",
			db: &DB{
				driverDB: &mock.Purger{
					PurgeFunc: func(_ context.Context, _ map[string][]string) (*driver.PurgeResult, error) {
						return nil, &Error{HTTPStatus: http.StatusNotImplemented, Message: "this feature is not yet implemented"}
					},
				},
			},
			status: StatusNotImplemented,
			err:    "this feature is not yet implemented",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Purge(context.Background(), test.docMap)
			testy.StatusError(t, test.err, test.status, err)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestBulkGet(t *testing.T) {
	type bulkGetTest struct {
		name    string
		db      *DB
		docs    []BulkGetReference
		options Options

		expected *Rows
		status   int
		err      string
	}

	tests := []bulkGetTest{
		{
			name:   "non-bulkGetter",
			db:     &DB{driverDB: &mock.DB{}},
			status: StatusNotImplemented,
			err:    "kivik: bulk get not supported by driver",
		},
		{
			name: "query error",
			db: &DB{driverDB: &mock.BulkGetter{
				BulkGetFunc: func(_ context.Context, docs []driver.BulkGetReference, opts map[string]interface{}) (driver.Rows, error) {
					return nil, errors.New("query error")
				},
			}},
			status: StatusInternalServerError,
			err:    "query error",
		},
		{
			name: "success",
			db: &DB{driverDB: &mock.BulkGetter{
				BulkGetFunc: func(_ context.Context, docs []driver.BulkGetReference, opts map[string]interface{}) (driver.Rows, error) {
					return &mock.Rows{ID: "bulkGet1"}, nil
				},
			}},
			expected: &Rows{
				iter: &iter{
					feed: &rowsIterator{
						Rows: &mock.Rows{ID: "bulkGet1"},
					},
					curVal: &driver.Row{},
				},
				rowsi: &mock.Rows{ID: "bulkGet1"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.BulkGet(context.Background(), test.docs, test.options)
			testy.StatusError(t, test.err, test.status, err)
			result.cancel = nil // Determinism
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestClientClose(t *testing.T) {
	type tst struct {
		db  *DB
		err string
	}
	tests := testy.NewTable()
	tests.Add("non-closer", tst{
		db: &DB{driverDB: &mock.DB{}},
	})
	tests.Add("error", tst{
		db: &DB{driverDB: &mock.DBCloser{
			CloseFunc: func(_ context.Context) error {
				return errors.New("close err")
			},
		}},
		err: "close err",
	})
	tests.Add("success", tst{
		db: &DB{driverDB: &mock.DBCloser{
			CloseFunc: func(_ context.Context) error {
				return nil
			},
		}},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		err := test.db.Close(context.Background())
		testy.Error(t, test.err, err)
	})
}

func TestRevsDiff(t *testing.T) {
	type tt struct {
		db       *DB
		revMap   interface{}
		status   int
		err      string
		expected interface{}
	}
	tests := testy.NewTable()
	tests.Add("non-DBReplicator", tt{
		db:     &DB{driverDB: &mock.DB{}},
		status: http.StatusNotImplemented,
		err:    "kivik: _revs_diff not supported by driver",
	})
	tests.Add("network error", tt{
		db: &DB{driverDB: &mock.RevsDiffer{
			RevsDiffFunc: func(_ context.Context, revMap interface{}) (driver.Rows, error) {
				return nil, errors.New("net error")
			},
		}},
		status: http.StatusInternalServerError,
		err:    "net error",
	})
	tests.Add("success", tt{
		db: &DB{driverDB: &mock.RevsDiffer{
			RevsDiffFunc: func(_ context.Context, revMap interface{}) (driver.Rows, error) {
				return &mock.Rows{ID: "a"}, nil
			},
		}},
		expected: &Rows{
			iter: &iter{
				feed: &rowsIterator{
					Rows: &mock.Rows{ID: "a"},
				},
				curVal: &driver.Row{},
			},
			rowsi: &mock.Rows{ID: "a"},
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		rows, err := tt.db.RevsDiff(context.Background(), tt.revMap)
		testy.StatusError(t, tt.err, tt.status, err)
		rows.cancel = nil // Determinism
		if d := diff.Interface(tt.expected, rows); d != nil {
			t.Error(d)
		}
	})
}
