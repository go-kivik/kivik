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

package sqlite

import (
	"math"
	"testing"
)

func Test_toInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    int64
		wantErr string
	}{
		{
			name:  "int64",
			input: int64(123),
			want:  123,
		},
		{
			name:    "uint64 overflow",
			input:   uint64(math.MaxUint64),
			wantErr: "error",
		},
		{
			name:  "large uint64 that does not overflow",
			input: uint64(math.MaxInt64),
			want:  math.MaxInt64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toInt64(tt.input, "error")
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != tt.wantErr {
				t.Errorf("Unexpected error, got %v want %v", errMsg, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got != tt.want {
				t.Errorf("Unexpected result, got %v want %v", got, tt.want)
			}
		})
	}
}
