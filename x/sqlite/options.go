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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/mango"
)

type optsMap map[string]any

func newOpts(options driver.Options) optsMap {
	opts := map[string]any{}
	options.Apply(opts)
	return opts
}

// get works like standard map access, but allows for multiple keys to be
// checked in order.
func (o optsMap) get(key ...string) (string, any, bool) {
	for _, k := range key {
		v, ok := o[k]
		if ok {
			return k, v, true
		}
	}
	return "", nil, false
}

func parseJSONKey(key string, in any) (string, error) {
	switch t := in.(type) {
	case json.RawMessage:
		var v any
		if err := json.Unmarshal(t, &v); err != nil {
			return "", &internal.Error{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid value for '%s': %w in key", key, err)}
		}
		return string(t), nil
	default:
		v, err := json.Marshal(t)
		if err != nil {
			return "", &internal.Error{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid value for '%s': %w in key", key, err)}
		}
		return string(v), nil
	}
}

func (o optsMap) endKey() (string, error) {
	key, value, ok := o.get("endkey", "end_key")
	if !ok {
		return "", nil
	}
	return parseJSONKey(key, value)
}

func (o optsMap) endkeyDocID() (string, error) {
	key, value, ok := o.get("endkey_docid", "end_key_doc_id")
	if !ok {
		return "", nil
	}
	v, ok := value.(string)
	if !ok {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for '%s': %v", key, value)}
	}
	return v, nil
}

func (o optsMap) startkeyDocID() (string, error) {
	key, value, ok := o.get("startkey_docid", "start_key_doc_id")
	if !ok {
		return "", nil
	}
	v, ok := value.(string)
	if !ok {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for '%s': %v", key, value)}
	}
	return v, nil
}

func (o optsMap) key() (string, error) {
	value, ok := o["key"]
	if !ok {
		return "", nil
	}
	return parseJSONKey("key", value)
}

func (o optsMap) keys() ([]string, error) {
	raw, ok := o["keys"]
	if !ok {
		return nil, nil
	}
	var tmp json.RawMessage
	switch t := raw.(type) {
	case json.RawMessage:
		tmp = t
	default:
		var err error
		tmp, err = json.Marshal(raw)
		if err != nil {
			return nil, &internal.Error{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid value for 'keys': %w", err)}
		}
	}
	var out []json.RawMessage
	if err := json.Unmarshal(tmp, &out); err != nil {
		return nil, &internal.Error{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid value for 'keys': %w", err)}
	}
	keys := make([]string, len(out))
	for i, v := range out {
		keys[i] = string(v)
	}
	return keys, nil
}

func (o optsMap) inclusiveEnd() (bool, error) {
	param, ok := o["inclusive_end"]
	if !ok {
		return true, nil
	}
	v, ok := toBool(param)
	if !ok {
		return false, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'inclusive_end': %v", param)}
	}
	return v, nil
}

func (o optsMap) startKey() (string, error) {
	key, value, ok := o.get("startkey", "start_key")
	if !ok {
		return "", nil
	}
	return parseJSONKey(key, value)
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
	parts := strings.SplitN(filter, "/", 2)
	if len(parts) != 2 {
		return "", "", "", &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf(`'%s' must be of the form 'designname/filtername'`, field)}
	}
	return filterType, "_design/" + parts[0], parts[1], nil
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

// limit returns the limit value as an int64, or -1 if the limit is unset.
// If the limit is invalid, an error is returned with status 400.
func (o optsMap) limit() (int64, error) {
	in, ok := o["limit"]
	if !ok {
		return -1, nil
	}
	return toInt64(in, "invalid value for 'limit'")
}

// skip returns the skip value as an int64, or 0 if the skip is unset.
// If the skip is invalid, an error is returned with status 400.
func (o optsMap) skip() (int64, error) {
	in, ok := o["skip"]
	if !ok {
		return 0, nil
	}
	return toInt64(in, "invalid value for 'skip'")
}

func (o optsMap) fields() ([]string, error) {
	raw, ok := o["fields"]
	if !ok {
		return nil, nil
	}

	f, ok := raw.([]any)
	if !ok {
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'fields': %v", raw)}
	}
	fields := make([]string, 0, len(f))
	for _, v := range f {
		s, ok := v.(string)
		if !ok {
			return nil, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid 'fields' field: %v", v)}
		}
		fields = append(fields, s)
	}
	return fields, nil
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

// toInt64 converts the input to a int64. If the input is malformed, it
// returns an error with msg as the message, and 400 as the status code.
func toInt64(in any, msg string) (int64, error) {
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
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("%s: %v", msg, in)}
		}
		return int64(t), nil
	case string:
		i, err := strconv.ParseInt(t, 10, 64)
		if err == nil {
			return i, nil
		}
		f, err := strconv.ParseFloat(t, 64)
		if err == nil {
			return int64(f), nil
		}
	case float32:
		i := int64(t)
		if float32(i) != t {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("%s: %v", msg, in)}
		}
		return i, nil
	case float64:
		i := int64(t)
		if float64(i) != t {
			return 0, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("%s: %v", msg, in)}
		}
		return i, nil
	}
	return 0, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("%s: %v", msg, in)}
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

func (o optsMap) sorted() (bool, error) {
	param, ok := o["sorted"]
	if !ok {
		return true, nil
	}
	v, ok := toBool(param)
	if !ok {
		return false, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'sorted': %v", param)}
	}
	if _, ok := o["descending"]; ok {
		// If descending is set to anything, then sorted must be true.
		// Error handling for invalid descending values is handled elsewhere.
		return true, nil
	}
	return v, nil
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
	return "", &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'update': %v", v)}
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
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'reduce': %v", raw)}
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
		return false, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'group': %v", raw)}
	}
	return v, nil
}

func (o optsMap) groupLevel() (uint64, error) {
	raw, ok := o["group_level"]
	if !ok {
		return 0, nil
	}
	return toUint64(raw, "invalid value for 'group_level'")
}

func (o optsMap) sort() ([]string, error) {
	raw, ok := o["sort"]
	if !ok {
		return nil, nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'sort': %v", raw)}
	}
	sort := make([]string, len(list))
	for i, v := range list {
		s, ok := v.(string)
		if !ok {
			return nil, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid 'sort' field: %v", v)}
		}
		sort[i] = s
	}
	return sort, nil
}

func (o optsMap) bookmark() (string, error) {
	raw, ok := o["bookmark"]
	if !ok {
		return "", nil
	}
	v, ok := raw.(string)
	if !ok {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'bookmark': %v", raw)}
	}
	bookmark, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'bookmark': %v", raw)}
	}
	return string(bookmark), nil
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

func (o optsMap) updateSeq() (bool, error) {
	param, ok := o["update_seq"]
	if !ok {
		return false, nil
	}
	v, ok := toBool(param)
	if !ok {
		return false, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'update_seq': %v", param)}
	}
	return v, nil
}

func (o optsMap) attEncodingInfo() (bool, error) {
	param, ok := o["att_encoding_info"]
	if !ok {
		return false, nil
	}
	v, ok := toBool(param)
	if !ok {
		return false, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'att_encoding_info': %v", param)}
	}
	return v, nil
}

const defaultWhereCap = 3

// buildReduceCacheWhere returns WHERE conditions for use when querying the
// reduce cache.
func (v viewOptions) buildReduceCacheWhere(args *[]any) []string {
	where := make([]string, 0, defaultWhereCap)
	if v.endkey != "" {
		op := endKeyOp(v.descending, v.inclusiveEnd)
		where = append(where, fmt.Sprintf("view.last_key %s $%d", op, len(*args)+1))
		*args = append(*args, v.endkey)
	}
	if v.startkey != "" {
		op := startKeyOp(v.descending)
		where = append(where, fmt.Sprintf("view.first_key %s $%d", op, len(*args)+1))
		*args = append(*args, v.startkey)
	}
	if v.key != "" {
		idx := strconv.Itoa(len(*args) + 1)
		where = append(where, "view.last_key = $"+idx, "view.first_key = $"+idx)
		*args = append(*args, v.key)
	}
	if len(v.keys) > 0 {
		for _, key := range v.keys {
			idx := strconv.Itoa(len(*args) + 1)
			where = append(where, "view.last_key = $"+idx, "view.first_key = $"+idx)
			*args = append(*args, key)
		}
	}
	return where
}

// buildGroupWhere returns WHERE conditions for use with grouping.
func (v viewOptions) buildGroupWhere(args *[]any) []string {
	where := make([]string, 0, defaultWhereCap)
	if v.endkey != "" {
		op := endKeyOp(v.descending, v.inclusiveEnd)
		where = append(where, fmt.Sprintf("view.key %s $%d", op, len(*args)+1))
		*args = append(*args, v.endkey)
	}
	if v.startkey != "" {
		op := startKeyOp(v.descending)
		where = append(where, fmt.Sprintf("view.key %s $%d", op, len(*args)+1))
		*args = append(*args, v.startkey)
	}
	if v.key != "" {
		where = append(where, "view.key = $"+strconv.Itoa(len(*args)+1))
		*args = append(*args, v.key)
	}
	if len(v.keys) > 0 {
		where = append(where, fmt.Sprintf("view.key IN (%s)", placeholders(len(v.keys), len(*args)+1)))
		for _, key := range v.keys {
			*args = append(*args, key)
		}
	}
	return where
}

func (v viewOptions) bookmarkWhere() string {
	if v.bookmark != "" {
		return `WHERE main.doc_number > (SELECT doc_number FROM bookmark)`
	}
	return ""
}

// buildWhere returns WHERE conditions based on the provided configuration
// arguments, and may append to args as needed.
func (v viewOptions) buildWhere(args *[]any) []string {
	where := make([]string, 0, defaultWhereCap)
	switch v.view {
	case viewAllDocs:
		where = append(where, `view.key NOT LIKE '"_local/%'`)
	case viewLocalDocs:
		where = append(where, `view.key LIKE '"_local/%'`)
	case viewDesignDocs:
		where = append(where, `view.key LIKE '"_design/%'`)
	}
	if v.endkey != "" {
		op := endKeyOp(v.descending, v.inclusiveEnd)
		where = append(where, fmt.Sprintf("view.key %s $%d", op, len(*args)+1))
		*args = append(*args, v.endkey)
		if v.endkeyDocID != "" {
			where = append(where, fmt.Sprintf("view.id %s $%d", op, len(*args)+1))
			*args = append(*args, v.endkeyDocID)
		}
	}
	if v.startkey != "" {
		op := startKeyOp(v.descending)
		where = append(where, fmt.Sprintf("view.key %s $%d", op, len(*args)+1))
		*args = append(*args, v.startkey)
		if v.startkeyDocID != "" {
			where = append(where, fmt.Sprintf("view.id %s $%d", op, len(*args)+1))
			*args = append(*args, v.startkeyDocID)
		}
	}
	if v.key != "" {
		where = append(where, "view.key = $"+strconv.Itoa(len(*args)+1))
		*args = append(*args, v.key)
	}
	if len(v.keys) > 0 {
		where = append(where, fmt.Sprintf("view.key IN (%s)", placeholders(len(v.keys), len(*args)+1)))
		for _, key := range v.keys {
			*args = append(*args, key)
		}
	}
	return where
}

func (v viewOptions) buildOrderBy(moreColumns ...string) string {
	if v.sorted {
		direction := descendingToDirection(v.descending)
		conditions := make([]string, 0, len(moreColumns)+1)
		for _, col := range append([]string{"key"}, moreColumns...) {
			conditions = append(conditions, "view."+col+" "+direction)
		}
		return "ORDER BY " + strings.Join(conditions, ", ")
	}
	return ""
}

// reduceGroupLevel returns the calculated groupLevel value to pass to
// [github.com/go-kivik/kivik/v4/x/sqlite/reduce.Reduce].
//
//	-1: Maximum grouping, same as group=true
//	 0: No grouping, same as group=false
//	1+: Group by the first N elements of the key, same as group_level=N
func (v viewOptions) reduceGroupLevel() int {
	if v.groupLevel == 0 && v.group {
		return -1
	}
	return int(v.groupLevel)
}

// viewOptions are all of the options recognized by the view endpoints
// _design/<ddoc>/_view/<view>, _all_docs, _design_docs, and _local_docs.
//
// See https://docs.couchdb.org/en/stable/api/ddoc/views.html#api-ddoc-view
type viewOptions struct {
	view            string
	limit           int64
	skip            int64
	descending      bool
	includeDocs     bool
	conflicts       bool
	reduce          *bool
	group           bool
	groupLevel      uint64
	endkey          string
	startkey        string
	inclusiveEnd    bool
	attachments     bool
	update          string
	updateSeq       bool
	endkeyDocID     string
	startkeyDocID   string
	key             string
	keys            []string
	sorted          bool
	attEncodingInfo bool

	// Find-specific options
	selector  *mango.Selector
	findLimit int64
	findSkip  int64
	fields    []string
	bookmark  string
	sort      []string
}

// findOptions converts a _find query body into a viewOptions struct.
func findOptions(query any) (*viewOptions, error) {
	input := query.(json.RawMessage)
	var s struct {
		Selector *mango.Selector `json:"selector"`
	}
	if err := json.Unmarshal(input, &s); err != nil {
		return nil, &internal.Error{Status: http.StatusBadRequest, Err: err}
	}
	if s.Selector == nil {
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: "selector cannot be null"}
	}
	var o optsMap
	if err := json.Unmarshal(input, &o); err != nil {
		return nil, &internal.Error{Status: http.StatusBadRequest, Err: err}
	}

	limit, err := o.limit()
	if err != nil {
		return nil, err
	}
	skip, err := o.skip()
	if err != nil {
		return nil, err
	}
	fields, err := o.fields()
	if err != nil {
		return nil, err
	}
	conflicts, err := o.conflicts()
	if err != nil {
		return nil, err
	}
	bookmark, err := o.bookmark()
	if err != nil {
		return nil, err
	}
	if bookmark != "" {
		skip = 0
	}
	sort, err := o.sort()
	if err != nil {
		return nil, err
	}
	if len(sort) > 0 {
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: "no index exists for this sort, try indexing by the sort fields"}
	}

	v := &viewOptions{
		view:        viewAllDocs,
		conflicts:   conflicts,
		includeDocs: true,
		limit:       -1,
		findLimit:   limit,
		findSkip:    skip,
		selector:    s.Selector,
		fields:      fields,
		bookmark:    bookmark,
		sort:        sort,
	}

	return v, v.validate()
}

func (o optsMap) viewOptions(view string) (*viewOptions, error) {
	limit, err := o.limit()
	if err != nil {
		return nil, err
	}
	skip, err := o.skip()
	if err != nil {
		return nil, err
	}
	reduce, err := o.reduce()
	if err != nil {
		return nil, err
	}
	group, err := o.group()
	if err != nil {
		return nil, err
	}
	groupLevel, err := o.groupLevel()
	if err != nil {
		return nil, err
	}
	if isBuiltinView(view) {
		if groupLevel > 0 {
			return nil, &internal.Error{Status: http.StatusBadRequest, Message: "group_level is invalid for map-only views"}
		}
		if group {
			return nil, &internal.Error{Status: http.StatusBadRequest, Message: "group is invalid for map-only views"}
		}
	}
	conflicts, err := o.conflicts()
	if err != nil {
		return nil, err
	}
	descending, err := o.descending()
	if err != nil {
		return nil, err
	}
	endkey, err := o.endKey()
	if err != nil {
		return nil, err
	}
	startkey, err := o.startKey()
	if err != nil {
		return nil, err
	}
	includeDocs, err := o.includeDocs()
	if err != nil {
		return nil, err
	}
	attachments, err := o.attachments()
	if err != nil {
		return nil, err
	}
	inclusiveEnd, err := o.inclusiveEnd()
	if err != nil {
		return nil, err
	}
	update, err := o.update()
	if err != nil {
		return nil, err
	}
	updateSeq, err := o.updateSeq()
	if err != nil {
		return nil, err
	}
	endkeyDocID, err := o.endkeyDocID()
	if err != nil {
		return nil, err
	}
	startkeyDocID, err := o.startkeyDocID()
	if err != nil {
		return nil, err
	}
	key, err := o.key()
	if err != nil {
		return nil, err
	}
	keys, err := o.keys()
	if err != nil {
		return nil, err
	}
	if len(keys) > 0 && (key != "" || endkey != "" || startkey != "") {
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: "`keys` is incompatible with `key`, `start_key` and `end_key`"}
	}
	sorted, err := o.sorted()
	if err != nil {
		return nil, err
	}
	attEncodingInfo, err := o.attEncodingInfo()
	if err != nil {
		return nil, err
	}

	v := &viewOptions{
		view:            view,
		limit:           limit,
		skip:            skip,
		descending:      descending,
		includeDocs:     includeDocs,
		conflicts:       conflicts,
		reduce:          reduce,
		group:           group,
		groupLevel:      groupLevel,
		endkey:          endkey,
		startkey:        startkey,
		inclusiveEnd:    inclusiveEnd,
		attachments:     attachments,
		update:          update,
		updateSeq:       updateSeq,
		endkeyDocID:     endkeyDocID,
		startkeyDocID:   startkeyDocID,
		key:             key,
		keys:            keys,
		sorted:          sorted,
		attEncodingInfo: attEncodingInfo,
	}
	return v, v.validate()
}

func (v viewOptions) validate() error {
	descendingModifier := 1
	if v.descending {
		descendingModifier = -1
	}
	if v.endkey != "" && v.startkey != "" && couchdbCmpString(v.startkey, v.endkey)*descendingModifier > 0 {
		return &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("no rows can match your key range, reverse your start_key and end_key or set descending=%v", !v.descending)}
	}
	if v.key != "" {
		startFail := v.startkey != "" && couchdbCmpString(v.key, v.startkey)*descendingModifier < 0
		endFail := v.endkey != "" && couchdbCmpString(v.key, v.endkey)*descendingModifier > 0
		if startFail && v.endkey != "" || endFail && v.startkey != "" {
			return &internal.Error{Status: http.StatusBadRequest, Message: "no rows can match your key range, change your start_key, end_key, or key"}
		}
		if startFail {
			return &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("no rows can match your key range, change your start_key or key or set descending=%v", !v.descending)}
		}
		if endFail {
			return &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("no rows can match your key range, reverse your end_key or key or set descending=%v", !v.descending)}
		}
	}

	return nil
}
