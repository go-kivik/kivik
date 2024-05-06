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
	"net/http"
	"strconv"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/x/sqlite/v4/internal"
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

// limit returns 0 if the limit is unset, or the limit value as a uint64. An
// explicit limit of 0 is converted to 1, as per CouchDB docs.
func (o optsMap) limit() (uint64, error) {
	in, ok := o["limit"]
	if !ok {
		return 0, nil
	}
	limit, err := toUint64(in, "malformed 'limit' parameter")
	if err != nil {
		return 0, err
	}
	if limit == 0 {
		limit = 1
	}
	return limit, nil
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

// toInt64 converts the input to a int64. If the input is malformed, it
// returns an error with msg as the message, and 400 as the status code.
func toInt64(in interface{}, msg string) (int64, error) {
	switch t := in.(type) {
	case int:
		return int64(t), nil
	case int64:
		return t, nil
	case int8:
		return int64(t), nil
	case int16:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case uint:
		return int64(t), nil
	case uint8:
		return int64(t), nil
	case uint16:
		return int64(t), nil
	case uint32:
		return int64(t), nil
	case uint64:
		if t > math.MaxInt64 {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: msg}
		}
		return int64(t), nil
	case string:
		i, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: msg}
		}
		return i, nil
	case float32:
		i := int64(t)
		if float32(i) != t {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: msg}
		}
		return i, nil
	case float64:
		i := int64(t)
		if float64(i) != t {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: msg}
		}
		return i, nil
	default:
		return 0, &internal.Error{Status: http.StatusBadRequest, Message: msg}
	}
}

// toFloat64 converts the input to a float64. If the input is malformed, it
// returns an error with msg as the message, and 400 as the status code.
func toFloat64(in interface{}) (float64, bool) {
	switch t := in.(type) {
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true
	case string:
		i, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0, false
		}
		return i, true
	case float32:
		return float64(t), true
	case float64:
		return t, true
	default:
		return 0, false
	}
}

func toBool(in interface{}) (value bool, ok bool) {
	switch t := in.(type) {
	case bool:
		return t, true
	case string:
		b, err := strconv.ParseBool(t)
		return b, err == nil
	default:
		return false, false
	}
}

func (o optsMap) direction() string {
	if v, _ := toBool(o["descending"]); v {
		return "DESC"
	}
	return "ASC"
}

func (o optsMap) includeDocs() bool {
	v, _ := toBool(o["include_docs"])
	return v
}

func (o optsMap) attachments() bool {
	v, _ := toBool(o["attachments"])
	return v
}

func (o optsMap) latest() bool {
	v, _ := toBool(o["latest"])
	return v
}

func (o optsMap) revs() bool {
	v, _ := toBool(o["revs"])
	return v
}

const (
	updateModeTrue  = "true"
	updateModeFalse = "false"
	updateModeLazy  = "lazy"
)

func (o optsMap) update() (string, error) {
	v, ok := o["update"]
	if !ok {
		return updateModeTrue, nil
	}
	switch t := v.(type) {
	case bool:
		if t {
			return updateModeTrue, nil
		}
		return updateModeFalse, nil
	case string:
		switch t {
		case "true":
			return updateModeTrue, nil
		case "false":
			return updateModeFalse, nil
		case "lazy":
			return updateModeLazy, nil
		}
	}
	return "", &internal.Error{Status: http.StatusBadRequest, Message: "invalid value for `update`"}
}

func (o optsMap) reduce() (*bool, error) {
	if group, _ := o.group(); group {
		return &group, nil
	}
	raw, ok := o["reduce"]
	if !ok {
		return nil, nil
	}
	v, ok := toBool(raw)
	if !ok {
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: "invalid value for `reduce`"}
	}
	return &v, nil
}

func (o optsMap) group() (bool, error) {
	if groupLevel, _ := o.groupLevel(); groupLevel > 0 {
		return groupLevel > 0, nil
	}
	raw, ok := o["group"]
	if !ok {
		return false, nil
	}
	v, ok := toBool(raw)
	if !ok {
		return false, &internal.Error{Status: http.StatusBadRequest, Message: "invalid value for `group`"}
	}
	return v, nil
}

func (o optsMap) groupLevel() (uint64, error) {
	raw, ok := o["group_level"]
	if !ok {
		return 0, nil
	}
	return toUint64(raw, "invalid value for `group_level`")
}
