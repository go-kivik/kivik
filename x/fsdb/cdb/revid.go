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

package cdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// RevID is a CouchDB document revision identifier.
type RevID struct {
	Seq      int64
	Sum      string
	original string
}

// Changed returns true if the rev has changed since being read.
func (r *RevID) Changed() bool {
	return r.String() != r.original
}

// UnmarshalText xatisfies the json.Unmarshaler interface.
func (r *RevID) UnmarshalText(p []byte) error {
	r.original = string(p)
	if bytes.Contains(p, []byte("-")) {
		parts := bytes.SplitN(p, []byte("-"), 2)
		seq, err := strconv.ParseInt(string(parts[0]), 10, 64)
		if err != nil {
			return err
		}
		r.Seq = seq
		if len(parts) > 1 {
			r.Sum = string(parts[1])
		}
		return nil
	}
	r.Sum = ""
	seq, err := strconv.ParseInt(string(p), 10, 64)
	if err != nil {
		return err
	}
	r.Seq = seq
	return nil
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (r *RevID) UnmarshalJSON(p []byte) error {
	if p[0] == '"' {
		var str string
		if e := json.Unmarshal(p, &str); e != nil {
			return e
		}
		r.original = str
		parts := strings.SplitN(str, "-", 2)
		seq, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return err
		}
		r.Seq = seq
		if len(parts) > 1 {
			r.Sum = parts[1]
		}
		return nil
	}
	r.original = string(p)
	r.Sum = ""
	return json.Unmarshal(p, &r.Seq)
}

// MarshalText satisfies the encoding.TextMarshaler interface.
func (r RevID) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r RevID) String() string {
	if r.Seq == 0 {
		return ""
	}
	return fmt.Sprintf("%d-%s", r.Seq, r.Sum)
}

// IsZero returns true if r is uninitialized.
func (r *RevID) IsZero() bool {
	return r.Seq == 0
}

// Equal returns true if rev and revid are equal.
func (r RevID) Equal(revid RevID) bool {
	return r.Seq == revid.Seq && r.Sum == revid.Sum
}
