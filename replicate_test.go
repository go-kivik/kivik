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

package xkivik

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	kivikmock "github.com/go-kivik/kivik/v4/mockdb"
	_ "github.com/go-kivik/kivik/v4/x/fsdb" // The filesystem driver
)

func TestReplicateMock(t *testing.T) {
	type tt struct {
		mockT, mockS   *kivikmock.Client
		target, source *kivik.DB
		options        kivik.Option
		status         int
		err            string
		result         *ReplicationResult
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
			result: &ReplicationResult{},
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
			result: &ReplicationResult{},
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
		sdb.ExpectGet().
			WithDocID("foo").
			WithOptions(kivik.Params(map[string]interface{}{
				"rev":         "2-7051cbe5c8faecd085a3fa619e6e6337",
				"revs":        true,
				"attachments": true,
			})).
			WillReturn(kivikmock.NewRows().AddRow(&driver.Row{
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
			result: &ReplicationResult{
				DocsRead:       1,
				DocsWritten:    1,
				MissingChecked: 1,
				MissingFound:   1,
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		result, err := Replicate(context.TODO(), tt.target, tt.source, tt.options)
		testy.StatusError(t, tt.err, tt.status, err)
		if tt.mockT != nil {
			testy.Error(t, "", tt.mockT.ExpectationsWereMet())
		}
		if tt.mockS != nil {
			testy.Error(t, "", tt.mockS.ExpectationsWereMet())
		}
		result.StartTime = time.Time{}
		result.EndTime = time.Time{}
		if d := testy.DiffAsJSON(tt.result, result); d != nil {
			t.Error(d)
		}
	})
}

func TestReplicate(t *testing.T) {
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
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
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
		result, err := Replicate(context.TODO(), tt.target, tt.source, tt.options)
		testy.StatusError(t, tt.err, tt.status, err)
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
