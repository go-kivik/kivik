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
	"net/http"
	"strconv"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

type optsMap map[string]interface{}

func newOpts(options driver.Options) optsMap {
	opts := map[string]interface{}{}
	options.Apply(opts)
	return opts
}

func (o optsMap) endKey() string {
	if endkey, ok := o["endkey"].(string); ok {
		return endkey
	}
	if endkey, ok := o["end_key"].(string); ok {
		return endkey
	}
	return ""
}

func (o optsMap) inclusiveEnd() bool {
	inclusiveEnd, ok := o["inclusive_end"].(bool)
	return !ok || inclusiveEnd
}

func (o optsMap) startKey() string {
	if startkey, ok := o["startkey"].(string); ok {
		return startkey
	}
	if startkey, ok := o["start_key"].(string); ok {
		return startkey
	}
	return ""
}

func (o optsMap) rev() string {
	rev, _ := o["rev"].(string)
	return rev
}

func (o optsMap) newEdits() bool {
	newEdits, ok := o["new_edits"].(bool)
	if !ok {
		return true
	}
	return newEdits
}

func (o optsMap) feed() (string, error) {
	feed, ok := o["feed"].(string)
	if !ok {
		return "normal", nil
	}
	switch feed {
	case feedNormal, feedLongpoll:
		return feed, nil
	}
	return "", &internal.Error{Status: http.StatusBadRequest, Message: "supported `feed` types: normal, longpoll"}
}

// since returns true if the value is "now", otherwise it returns the sequence
// id as a uint64.
func (o optsMap) since() (bool, *uint64, error) {
	in, ok := o["since"].(string)
	if !ok {
		return false, nil, nil
	}
	if in == "now" {
		return true, nil, nil
	}
	since, err := toUint64(in, "malformed sequence supplied in 'since' parameter")
	return false, &since, err
}

func (o optsMap) limit() (*uint64, error) {
	in, ok := o["limit"]
	if !ok {
		return nil, nil
	}
	limit, err := toUint64(in, "malformed 'limit' parameter")
	if err != nil {
		return nil, err
	}
	if limit == 0 {
		limit = 1
	}
	return &limit, nil
}

// toUint64 converts the input to a uint64. If the input is malformed, it
// returns an error with msg as the message, and 400 as the status code.
func toUint64(in interface{}, msg string) (uint64, error) {
	checkSign := func(i int64) (uint64, error) {
		if i < 0 {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: msg}
		}
		return uint64(i), nil
	}
	switch t := in.(type) {
	case int:
		return checkSign(int64(t))
	case int64:
		return checkSign(t)
	case int8:
		return checkSign(int64(t))
	case int16:
		return checkSign(int64(t))
	case int32:
		return checkSign(int64(t))
	case uint:
		return uint64(t), nil
	case uint8:
		return uint64(t), nil
	case uint16:
		return uint64(t), nil
	case uint32:
		return uint64(t), nil
	case uint64:
		return t, nil
	case string:
		i, err := strconv.ParseUint(t, 10, 64)
		if err != nil {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: msg}
		}
		return i, nil
	case float32:
		i := uint64(t)
		if float32(i) != t {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: msg}
		}
		return i, nil
	case float64:
		i := uint64(t)
		if float64(i) != t {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: msg}
		}
		return i, nil
	default:
		return 0, &internal.Error{Status: http.StatusBadRequest, Message: msg}
	}
}

func (o optsMap) direction() string {
	if desc, ok := o["descending"]; ok {
		switch t := desc.(type) {
		case bool:
			if t {
				return "DESC"
			}
		case string:
			b, _ := strconv.ParseBool(t)
			if b {
				return "DESC"
			}
		}
	}
	return "ASC"
}
