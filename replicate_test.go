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

package kivik_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	kivikmock "github.com/go-kivik/kivik/v4/mockdb"
	_ "github.com/go-kivik/kivik/v4/x/fsdb" // The filesystem driver
)

const isGopherJS117 = runtime.GOARCH == "js"

func TestReplicateMock(t *testing.T) {
	type tt struct {
		mockT, mockS   *kivikmock.Client
		target, source *kivik.DB
		options        kivik.Option
		status         int
		err            string
		result         *kivik.ReplicationResult
	}
	tests := testy.NewTable()
	tests.Add("no changes", func(t *testing.T) interface{} {
		source, mock := kivikmock.NewT(t)
		db := mock.NewDB()
		mock.ExpectDB().WillReturn(db)
		db.ExpectChanges().WillReturn(kivikmock.NewChanges())

		return tt{
			mockS:  mock,
			source: source.DB("src"),
			result: &kivik.ReplicationResult{},
		}
	})
	tests.Add("up to date", func(t *testing.T) interface{} {
		source, smock := kivikmock.NewT(t)
		sdb := smock.NewDB()
		smock.ExpectDB().WillReturn(sdb)
		sdb.ExpectChanges().WillReturn(kivikmock.NewChanges().
			AddChange(&driver.Change{
				ID:      "foo",
				Changes: []string{"2-7051cbe5c8faecd085a3fa619e6e6337"},
				Seq:     "3-g1AAAAG3eJzLYWBg4MhgTmHgz8tPSTV0MDQy1zMAQsMcoARTIkOS_P___7MSGXAqSVIAkkn2IFUZzIkMuUAee5pRqnGiuXkKA2dpXkpqWmZeagpu_Q4g_fGEbEkAqaqH2sIItsXAyMjM2NgUUwdOU_JYgCRDA5ACGjQfn30QlQsgKvcjfGaQZmaUmmZClM8gZhyAmHGfsG0PICrBPmQC22ZqbGRqamyIqSsLAAArcXo",
			}))

		target, tmock := kivikmock.NewT(t)
		tdb := tmock.NewDB()
		tmock.ExpectDB().WillReturn(tdb)
		tdb.ExpectRevsDiff().
			WithRevLookup(map[string][]string{
				"foo": {"2-7051cbe5c8faecd085a3fa619e6e6337"},
			}).
			WillReturn(kivikmock.NewRows())

		return tt{
			mockS:  smock,
			mockT:  tmock,
			source: source.DB("src"),
			target: target.DB("tgt"),
			result: &kivik.ReplicationResult{},
		}
	})
	tests.Add("one update", func(t *testing.T) interface{} {
		source, smock := kivikmock.NewT(t)
		sdb := smock.NewDB()
		smock.ExpectDB().WillReturn(sdb)
		sdb.ExpectChanges().WillReturn(kivikmock.NewChanges().
			AddChange(&driver.Change{
				ID:      "foo",
				Changes: []string{"2-7051cbe5c8faecd085a3fa619e6e6337"},
				Seq:     "3-g1AAAAG3eJzLYWBg4MhgTmHgz8tPSTV0MDQy1zMAQsMcoARTIkOS_P___7MSGXAqSVIAkkn2IFUZzIkMuUAee5pRqnGiuXkKA2dpXkpqWmZeagpu_Q4g_fGEbEkAqaqH2sIItsXAyMjM2NgUUwdOU_JYgCRDA5ACGjQfn30QlQsgKvcjfGaQZmaUmmZClM8gZhyAmHGfsG0PICrBPmQC22ZqbGRqamyIqSsLAAArcXo",
			}))

		target, tmock := kivikmock.NewT(t)
		tdb := tmock.NewDB()
		tmock.ExpectDB().WillReturn(tdb)
		tdb.ExpectRevsDiff().
			WithRevLookup(map[string][]string{
				"foo": {"2-7051cbe5c8faecd085a3fa619e6e6337"},
			}).
			WillReturn(kivikmock.NewRows().
				AddRow(&driver.Row{
					ID:    "foo",
					Value: strings.NewReader(`{"missing":["2-7051cbe5c8faecd085a3fa619e6e6337"]}`),
				}))
		sdb.ExpectOpenRevs().WillReturnError(&internal.Error{Status: http.StatusNotImplemented})
		sdb.ExpectGet().
			WithDocID("foo").
			WithOptions(kivik.Params(map[string]interface{}{
				"rev":         "2-7051cbe5c8faecd085a3fa619e6e6337",
				"revs":        true,
				"attachments": true,
			})).
			WillReturn(&driver.Document{
				Body: io.NopCloser(strings.NewReader(`{"_id":"foo","_rev":"2-7051cbe5c8faecd085a3fa619e6e6337","foo":"bar"}`)),
			})
		tdb.ExpectPut().
			WithDocID("foo").
			WithOptions(kivik.Param("new_edits", false)).
			WillReturn("2-7051cbe5c8faecd085a3fa619e6e6337")

		return tt{
			mockS:  smock,
			mockT:  tmock,
			source: source.DB("src"),
			target: target.DB("tgt"),
			result: &kivik.ReplicationResult{
				DocsRead:       1,
				DocsWritten:    1,
				MissingChecked: 1,
				MissingFound:   1,
			},
		}
	})
	tests.Add("one update with OpenRevs", func(t *testing.T) interface{} {
		source, smock := kivikmock.NewT(t)
		sdb := smock.NewDB()
		smock.ExpectDB().WillReturn(sdb)
		sdb.ExpectChanges().WillReturn(kivikmock.NewChanges().
			AddChange(&driver.Change{
				ID:      "foo",
				Changes: []string{"2-7051cbe5c8faecd085a3fa619e6e6337"},
				Seq:     "3-g1AAAAG3eJzLYWBg4MhgTmHgz8tPSTV0MDQy1zMAQsMcoARTIkOS_P___7MSGXAqSVIAkkn2IFUZzIkMuUAee5pRqnGiuXkKA2dpXkpqWmZeagpu_Q4g_fGEbEkAqaqH2sIItsXAyMjM2NgUUwdOU_JYgCRDA5ACGjQfn30QlQsgKvcjfGaQZmaUmmZClM8gZhyAmHGfsG0PICrBPmQC22ZqbGRqamyIqSsLAAArcXo",
			}))

		target, tmock := kivikmock.NewT(t)
		tdb := tmock.NewDB()
		tmock.ExpectDB().WillReturn(tdb)
		tdb.ExpectRevsDiff().
			WithRevLookup(map[string][]string{
				"foo": {"2-7051cbe5c8faecd085a3fa619e6e6337"},
			}).
			WillReturn(kivikmock.NewRows().
				AddRow(&driver.Row{
					ID:    "foo",
					Value: strings.NewReader(`{"missing":["2-7051cbe5c8faecd085a3fa619e6e6337"]}`),
				}))
		sdb.ExpectOpenRevs().
			WithDocID("foo").
			WillReturn(kivikmock.NewRows().AddRow(&driver.Row{
				ID:  "foo",
				Rev: "2-7051cbe5c8faecd085a3fa619e6e6337",
				Doc: strings.NewReader(`{"_id":"foo","_rev":"2-7051cbe5c8faecd085a3fa619e6e6337","foo":"bar"}`),
			}))
		tdb.ExpectPut().
			WithDocID("foo").
			WithOptions(kivik.Param("new_edits", false)).
			WillReturn("2-7051cbe5c8faecd085a3fa619e6e6337")

		return tt{
			mockS:  smock,
			mockT:  tmock,
			source: source.DB("src"),
			target: target.DB("tgt"),
			result: &kivik.ReplicationResult{
				DocsRead:       1,
				DocsWritten:    1,
				MissingChecked: 1,
				MissingFound:   1,
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		result, err := kivik.Replicate(context.TODO(), tt.target, tt.source, tt.options)
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if tt.mockT != nil {
			if err := tt.mockT.ExpectationsWereMet(); !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		}
		if tt.mockS != nil {
			if err := tt.mockS.ExpectationsWereMet(); !testy.ErrorMatches("", err) {
				t.Errorf("Unexpected error: %s", err)
			}
		}
		result.StartTime = time.Time{}
		result.EndTime = time.Time{}
		if d := testy.DiffAsJSON(tt.result, result); d != nil {
			t.Error(d)
		}
	})
}

func TestReplicate_with_callback(t *testing.T) {
	source, smock := kivikmock.NewT(t)
	sdb := smock.NewDB()
	smock.ExpectDB().WillReturn(sdb)
	sdb.ExpectChanges().WillReturn(kivikmock.NewChanges().
		AddChange(&driver.Change{
			ID:      "foo",
			Changes: []string{"2-7051cbe5c8faecd085a3fa619e6e6337"},
			Seq:     "3-g1AAAAG3eJzLYWBg4MhgTmHgz8tPSTV0MDQy1zMAQsMcoARTIkOS_P___7MSGXAqSVIAkkn2IFUZzIkMuUAee5pRqnGiuXkKA2dpXkpqWmZeagpu_Q4g_fGEbEkAqaqH2sIItsXAyMjM2NgUUwdOU_JYgCRDA5ACGjQfn30QlQsgKvcjfGaQZmaUmmZClM8gZhyAmHGfsG0PICrBPmQC22ZqbGRqamyIqSsLAAArcXo",
		}))

	target, tmock := kivikmock.NewT(t)
	tdb := tmock.NewDB()
	tmock.ExpectDB().WillReturn(tdb)
	tdb.ExpectRevsDiff().
		WithRevLookup(map[string][]string{
			"foo": {"2-7051cbe5c8faecd085a3fa619e6e6337"},
		}).
		WillReturn(kivikmock.NewRows().
			AddRow(&driver.Row{
				ID:    "foo",
				Value: strings.NewReader(`{"missing":["2-7051cbe5c8faecd085a3fa619e6e6337"]}`),
			}))
	sdb.ExpectOpenRevs().WillReturnError(&internal.Error{Status: http.StatusNotImplemented})
	sdb.ExpectGet().
		WithDocID("foo").
		WithOptions(kivik.Params(map[string]interface{}{
			"rev":         "2-7051cbe5c8faecd085a3fa619e6e6337",
			"revs":        true,
			"attachments": true,
		})).
		WillReturn(&driver.Document{
			Body: io.NopCloser(strings.NewReader(`{"_id":"foo","_rev":"2-7051cbe5c8faecd085a3fa619e6e6337","foo":"bar"}`)),
		})
	tdb.ExpectPut().
		WithDocID("foo").
		WithOptions(kivik.Param("new_edits", false)).
		WillReturn("2-7051cbe5c8faecd085a3fa619e6e6337")

	events := []kivik.ReplicationEvent{}

	_, err := kivik.Replicate(context.TODO(), target.DB("tgt"), source.DB("src"), kivik.ReplicateCallback(func(e kivik.ReplicationEvent) {
		events = append(events, e)
	}))
	if err != nil {
		t.Fatal(err)
	}

	expected := []kivik.ReplicationEvent{
		{
			Type: "changes",
			Read: true,
		},
		{
			Type:    "change",
			Read:    true,
			DocID:   "foo",
			Changes: []string{"2-7051cbe5c8faecd085a3fa619e6e6337"},
		},
		{
			Type: "revsdiff",
			Read: true,
		},
		{
			Type:  "revsdiff",
			Read:  true,
			DocID: "foo",
		},
		{
			Type:  "document",
			Read:  true,
			DocID: "foo",
		},
		{
			Type:  "document",
			DocID: "foo",
		},
	}
	if d := cmp.Diff(expected, events); d != "" {
		t.Error(d)
	}
}

func TestReplicate(t *testing.T) {
	if isGopherJS117 {
		t.Skip("Replication doesn't work in GopherJS 1.17")
	}
	type tt struct {
		path           string
		target, source *kivik.DB
		options        kivik.Option
		status         int
		err            string
	}
	tests := testy.NewTable()
	tests.Add("fs to fs", func(t *testing.T) interface{} {
		tmpdir := testy.CopyTempDir(t, "testdata/db4", 1)
		t.Cleanup(func() {
			_ = os.RemoveAll(tmpdir)
		})

		client, err := kivik.New("fs", tmpdir)
		if err != nil {
			t.Fatal(err)
		}
		if err := client.CreateDB(context.TODO(), "target"); err != nil {
			t.Fatal(err)
		}

		return tt{
			path:   tmpdir,
			target: client.DB("target"),
			source: client.DB("db4"),
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		result, err := kivik.Replicate(context.TODO(), tt.target, tt.source, tt.options)
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		result.StartTime = time.Time{}
		result.EndTime = time.Time{}
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t, "fs"), testy.JSONDir{
			Path:           tt.path,
			FileContent:    true,
			MaxContentSize: 100,
		}); d != nil {
			t.Error(d)
		}
	})
}
