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
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

type optsMap map[string]any

func newOpts(options driver.Options) optsMap {
	opts := map[string]any{}
	options.Apply(opts)
	return opts
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
	case feedNormal, feedLongpoll, feedContinuous:
		return feed, nil
	}
	return "", &internal.Error{Status: http.StatusBadRequest, Message: "supported `feed` types: normal, longpoll, continuous"}
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

// changesLimit returns the changesLimit value as a uint64, or 0 if the changesLimit is unset. An
// explicit changesLimit of 0 is converted to 1, as per CouchDB docs.
func (o optsMap) changesLimit() (uint64, error) {
	in, ok := o["limit"]
	if !ok {
		return 0, nil
	}
	limit, err := toUint64(in, "invalid value for 'limit'")
	if err != nil {
		return 0, err
	}
	if limit == 0 {
		limit = 1
	}
	return limit, nil
}

func (o optsMap) changesFilter() (filterType, filterDdoc, filterName string, _ error) {
	raw, ok := o["filter"]
	if !ok {
		return "", "", "", nil
	}
	filter, _ := raw.(string)
	field, filterType := "filter", "filter"
	switch filter {
	case "_doc_ids":
		return "_doc_ids", "", "", nil
	case "_view":
		raw, ok := o["view"]
		if !ok {
			return "", "", "", &internal.Error{Status: http.StatusBadRequest, Message: "filter=_view requires 'view' parameter"}
		}
		filter, _ = raw.(string)
		field, filterType = "view", "map"
	}
	ddoc, name, ok := strings.Cut(filter, "/")
	if !ok {
		return "", "", "", &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf(`'%s' must be of the form 'designname/filtername'`, field)}
	}
	return filterType, "_design/" + ddoc, name, nil
}

func (o optsMap) changesWhere(args *[]any) (string, error) {
	filterType, _, _, err := o.changesFilter()
	if err != nil {
		return "", err
	}
	if filterType != "_doc_ids" {
		return "", nil
	}

	raw, ok := o["doc_ids"]
	if !ok {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: "filter=_doc_ids requires 'doc_ids' parameter"}
	}
	list, ok := raw.([]any)
	if !ok {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'doc_ids': %v", raw)}
	}
	start := len(*args)
	for _, v := range list {
		if _, ok := v.(string); !ok {
			return "", &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid 'doc_ids' field: %v", v)}
		}
		*args = append(*args, v)
	}

	return fmt.Sprintf("WHERE results.id IN (%s)", placeholders(start+1, len(*args)-start)), nil
}

// toUint64 converts the input to a uint64. If the input is malformed, it
// returns an error with msg as the message, and 400 as the status code.
func toUint64(in any, msg string) (uint64, error) {
	checkSign := func(i int64) (uint64, error) {
		if i < 0 {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("%s: %v", msg, in)}
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
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("%s: %v", msg, in)}
		}
		return i, nil
	case float32:
		i := uint64(t)
		if float32(i) != t {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("%s: %v", msg, in)}
		}
		return i, nil
	case float64:
		i := uint64(t)
		if float64(i) != t {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("%s: %v", msg, in)}
		}
		return i, nil
	default:
		return 0, &internal.Error{Status: http.StatusBadRequest, Message: msg}
	}
}

func toBool(in any) (value bool, ok bool) {
	switch t := in.(type) {
	case bool:
		return t, true
	case string:
		switch t {
		case "true":
			return true, true
		case "false":
			return false, true
		}
		return false, false
	default:
		return false, false
	}
}

func (o optsMap) descending() (bool, error) {
	param, ok := o["descending"]
	if !ok {
		return false, nil
	}
	v, ok := toBool(param)
	if !ok {
		return false, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'descending': %v", param)}
	}
	return v, nil
}

func (o optsMap) includeDocs() (bool, error) {
	param, ok := o["include_docs"]
	if !ok {
		return false, nil
	}
	v, ok := toBool(param)
	if !ok {
		return false, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'include_docs': %v", param)}
	}
	return v, nil
}

func (o optsMap) attachments() (bool, error) {
	param, ok := o["attachments"]
	if !ok {
		return false, nil
	}
	v, ok := toBool(param)
	if !ok {
		return false, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'attachments': %v", param)}
	}
	return v, nil
}

func (o optsMap) latest() bool {
	v, _ := toBool(o["latest"])
	return v
}

func (o optsMap) revs() bool {
	v, _ := toBool(o["revs"])
	return v
}

func (o optsMap) conflicts() (bool, error) {
	if o.meta() {
		return true, nil
	}
	param, ok := o["conflicts"]
	if !ok {
		return false, nil
	}
	v, ok := toBool(param)
	if !ok {
		return false, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'conflicts': %v", param)}
	}
	return v, nil
}

func (o optsMap) meta() bool {
	v, _ := toBool(o["meta"])
	return v
}

func (o optsMap) deletedConflicts() bool {
	if o.meta() {
		return true
	}
	v, _ := toBool(o["deleted_conflicts"])
	return v
}

func (o optsMap) revsInfo() bool {
	if o.meta() {
		return true
	}
	v, _ := toBool(o["revs_info"])
	return v
}

func (o optsMap) localSeq() bool {
	v, _ := toBool(o["local_seq"])
	return v
}

func (o optsMap) attsSince() []string {
	attsSince, _ := o["atts_since"].([]string)
	return attsSince
}
