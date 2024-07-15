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

package ast

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"
)

var cmpOpts = cmp.AllowUnexported(unarySelector{}, combinationSelector{}, conditionSelector{})

func TestParse(t *testing.T) {
	type test struct {
		input   string
		want    Selector
		wantErr string
	}

	tests := testy.NewTable()
	tests.Add("empty", test{
		input: "{}",
		want: &combinationSelector{
			op:  OpAnd,
			sel: nil,
		},
	})
	tests.Add("implicit equality", test{
		input: `{"foo": "bar"}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpEqual,
			value: "bar",
		},
	})
	tests.Add("explicit equality", test{
		input: `{"foo": {"$eq": "bar"}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpEqual,
			value: "bar",
		},
	})
	tests.Add("explicit equality with too many object keys", test{
		input:   `{"foo": {"$eq": "bar", "$ne": "baz"}}`,
		wantErr: "too many keys in object",
	})
	tests.Add("implicit equality with empty object", test{
		input: `{"foo": {}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpEqual,
			value: map[string]interface{}{},
		},
	})
	tests.Add("explicit invalid comparison operator", test{
		input:   `{"foo": {"$invalid": "bar"}}`,
		wantErr: "invalid operator $invalid",
	})
	tests.Add("explicit equiality against object", test{
		input: `{"foo": {"$eq": {"bar": "baz"}}}`,
		want: &conditionSelector{
			field: "foo",
			op:    OpEqual,
			value: map[string]interface{}{"bar": "baz"},
		},
	})

	/*
		TODO:
		- $lt
		- $lte
		- $eq
		- $ne
		- $gt
		- $gte
		- $exists
		- $type
		- $in
		- $nin
		- $size
		- $mod
		- $regex
		- implicit $and
		- $and
		- $or
		- $not
		- $nor
		- $all
		- $elemMatch
		- $allMatch
		- $keyMapMatch


	*/

	tests.Run(t, func(t *testing.T, tt test) {
		got, err := Parse([]byte(tt.input))
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Fatalf("Unexpected error: %s", err)
		}
		if d := cmp.Diff(tt.want, got, cmpOpts); d != "" {
			t.Errorf("Unexpected result (-want +got):\n%s", d)
		}
	})
}
