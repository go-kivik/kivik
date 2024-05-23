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

//go:build !js
// +build !js

package sqlite

import (
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
)

type testReduceRows []*reduceRow

var _ reduceRows = &testReduceRows{}

func (r *testReduceRows) Next() (*reduceRow, error) {
	if len(*r) == 0 {
		return nil, io.EOF
	}
	row := (*r)[0]
	*r = (*r)[1:]
	return row, nil
}

func Test_reduceRows(t *testing.T) {
	t.Parallel()

	type test struct {
		rows          reduceRows
		reduceFuncJS  string
		vopts         *viewOptions
		want          []reducedRow
		wantErr       string
		wantErrStatus int
	}

	tests := testy.NewTable()
	tests.Add("no rows", test{
		rows: &testReduceRows{},
		want: []reducedRow{},
	})
	tests.Add("one row", test{
		rows: &testReduceRows{
			{ID: "foo", Key: "foo", Value: &[]string{"1"}[0]},
		},
		reduceFuncJS: `_sum`,
		want: []reducedRow{
			{
				Key:   "null",
				Value: "1",
			},
		},
	})
	tests.Add("two rows", test{
		rows: &testReduceRows{
			{ID: "foo", Key: "foo", Value: &[]string{"1"}[0]},
			{ID: "bar", Key: "bar", Value: &[]string{"1"}[0]},
		},
		reduceFuncJS: `_sum`,
		want: []reducedRow{
			{
				Key:   "null",
				Value: "2",
			},
		},
	})
	tests.Add("group=true", test{
		rows: &testReduceRows{
			{ID: "foo", Key: "foo", Value: &[]string{"1"}[0]},
			{ID: "bar", Key: "bar", Value: &[]string{"1"}[0]},
		},
		reduceFuncJS: `_sum`,
		vopts: &viewOptions{
			group: true,
		},
		want: []reducedRow{
			{
				Key:   "foo",
				Value: "1",
			},
			{
				Key:   "bar",
				Value: "1",
			},
		},
	})
	tests.Add("group_level=1", test{
		rows: &testReduceRows{
			{ID: "foo", Key: `["a","b"]`, Value: &[]string{"1"}[0]},
			{ID: "bar", Key: `["a","b"]`, Value: &[]string{"1"}[0]},
		},
		reduceFuncJS: `_sum`,
		vopts: &viewOptions{
			group:      true,
			groupLevel: 1,
		},
		want: []reducedRow{
			{
				Key:   `["a"]`,
				Value: "2",
			},
		},
	})

	/* TODO:
	- group_level=0 vs group_level = max
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()

		d := newDB(t)
		var reduceFuncJS *string
		if tt.reduceFuncJS != "" {
			reduceFuncJS = &tt.reduceFuncJS
		}
		vopts := &viewOptions{
			sorted: true,
		}
		if tt.vopts != nil {
			vopts = tt.vopts
		}
		got, err := d.DB.(*db).reduceRows(tt.rows, reduceFuncJS, vopts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantErrStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}
		checkReducedRows(t, tt.want, got)
	})
}

type reducedRow struct {
	ID    string
	Rev   string
	Key   string
	Value string
	Doc   string
	Error string
}

func checkReducedRows(t *testing.T, want []reducedRow, got *reducedRows) {
	t.Helper()
	g := make([]reducedRow, 0, len(*got))
	for _, row := range *got {
		newRow := reducedRow{
			ID:  row.ID,
			Rev: row.Rev,
			Key: string(row.Key),
		}
		if row.Value != nil {
			v, _ := io.ReadAll(row.Value)
			newRow.Value = string(v)
		}
		if row.Doc != nil {
			d, _ := io.ReadAll(row.Doc)
			newRow.Doc = string(d)
		}
		if row.Error != nil {
			newRow.Error = row.Error.Error()
		}
		g = append(g, newRow)
	}
	if d := cmp.Diff(want, g); d != "" {
		t.Errorf("Unexpected reduced rows:\n%s", d)
	}
}
