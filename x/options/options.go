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

// Package options implements options handling for standard CouchDB options.
// This package is not intended for public consumption, and the API is subject
// to change without warning.
package options

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/collate"
	"github.com/go-kivik/kivik/v4/x/mango"
)

// Map is a map of options.
type Map map[string]any

// New creates a new optsMap from the given driver.Options.
func New(options driver.Options) Map {
	opts := map[string]any{}
	if options != nil {
		options.Apply(opts)
	}
	return opts
}

// Get works like standard map access, but allows for multiple keys to be
// checked in order.
func (o Map) Get(key ...string) (string, any, bool) {
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

// EndKey returns the endkey option as an encoded JSON value.
func (o Map) EndKey() (string, error) {
	key, value, ok := o.Get("endkey", "end_key")
	if !ok {
		return "", nil
	}
	return parseJSONKey(key, value)
}

// EndKeyDocID returns the endkey_docid option.
func (o Map) EndKeyDocID() (string, error) {
	key, value, ok := o.Get("endkey_docid", "end_key_doc_id")
	if !ok {
		return "", nil
	}
	v, ok := value.(string)
	if !ok {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for '%s': %v", key, value)}
	}
	return v, nil
}

// StartKeyDocID returns the startkey_docid option.
func (o Map) StartKeyDocID() (string, error) {
	key, value, ok := o.Get("startkey_docid", "start_key_doc_id")
	if !ok {
		return "", nil
	}
	v, ok := value.(string)
	if !ok {
		return "", &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for '%s': %v", key, value)}
	}
	return v, nil
}

// Key returns the key option.
func (o Map) Key() (string, error) {
	value, ok := o["key"]
	if !ok {
		return "", nil
	}
	return parseJSONKey("key", value)
}

// Keys returns the keys option.
func (o Map) Keys() ([]string, error) {
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

// InclusiveEnd returns the inclusive_end option. The default is true.
func (o Map) InclusiveEnd() (bool, error) {
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

// StartKey returns the startkey option.
func (o Map) StartKey() (string, error) {
	key, value, ok := o.Get("startkey", "start_key")
	if !ok {
		return "", nil
	}
	return parseJSONKey(key, value)
}

// Rev returns the rev option.
func (o Map) Rev() string {
	rev, _ := o["rev"].(string)
	return rev
}

// NewEdits returns the new_edits option.
func (o Map) NewEdits() bool {
	newEdits, ok := o["new_edits"].(bool)
	if !ok {
		return true
	}
	return newEdits
}

// Feed returns the feed option.
func (o Map) Feed() (string, error) {
	feed, ok := o["feed"].(string)
	if !ok {
		return "normal", nil
	}
	switch feed {
	case FeedNormal, FeedLongpoll, FeedContinuous:
		return feed, nil
	}
	return "", &internal.Error{Status: http.StatusBadRequest, Message: "supported `feed` types: normal, longpoll, continuous"}
}

// Style returns the style option. Unrecognized values default to
// [StyleMainOnly].
func (o Map) Style() string {
	style, _ := o["style"].(string)
	switch style {
	case StyleAllDocs:
		return StyleAllDocs
	default:
		return StyleMainOnly
	}
}

// Timeout returns the timeout option as a time.Duration. The value is
// interpreted as milliseconds.
func (o Map) Timeout() (time.Duration, error) {
	in, ok := o["timeout"]
	if !ok {
		return 0, nil
	}
	ms, err := toUint64(in, "invalid value for 'timeout'")
	if err != nil {
		return 0, err
	}
	return time.Duration(ms) * time.Millisecond, nil
}

// Since returns true if the value is "now", otherwise it returns the sequence
// id as a uint64.
func (o Map) Since() (bool, *uint64, error) {
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

// ChangesLimit returns the ChangesLimit value as a uint64, or 0 if the ChangesLimit is unset. An
// explicit ChangesLimit of 0 is converted to 1, as per CouchDB docs.
func (o Map) ChangesLimit() (uint64, error) {
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

// ChangesFilter returns the filter options.
func (o Map) ChangesFilter() (filterType, filterDdoc, filterName string, _ error) {
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

// ChangesWhere returns the WHERE clause for the changes feed.
func (o Map) ChangesWhere(args *[]any) (string, error) {
	filterType, _, _, err := o.ChangesFilter()
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

// Limit returns the Limit value as an int64, or -1 if the Limit is unset.
// If the Limit is invalid, an error is returned with status 400.
func (o Map) Limit() (int64, error) {
	in, ok := o["limit"]
	if !ok {
		return -1, nil
	}
	return toInt64(in, "invalid value for 'limit'")
}

// Skip returns the Skip value as an int64, or 0 if the Skip is unset.
// If the Skip is invalid, an error is returned with status 400.
func (o Map) Skip() (int64, error) {
	in, ok := o["skip"]
	if !ok {
		return 0, nil
	}
	return toInt64(in, "invalid value for 'skip'")
}

// Fields returns the Fields option.
func (o Map) Fields() ([]string, error) {
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

// Sorted returns the sorted option.
func (o Map) Sorted() (bool, error) {
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

// Descending returns the descending option.
func (o Map) Descending() (bool, error) {
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

// IncludeDocs returns the include_docs option.
func (o Map) IncludeDocs() (bool, error) {
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

// Attachments returns the attachments option.
func (o Map) Attachments() (bool, error) {
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

// Latest returns the latest option.
func (o Map) Latest() bool {
	v, _ := toBool(o["latest"])
	return v
}

// Revs returns the revs option.
func (o Map) Revs() bool {
	v, _ := toBool(o["revs"])
	return v
}

// Update mode constants for view queries.
const (
	UpdateModeTrue  = "true"
	UpdateModeFalse = "false"
	UpdateModeLazy  = "lazy"
)

// Update returns the update option, which may be "true", "false", or "lazy".
func (o Map) Update() (string, error) {
	v, ok := o["update"]
	if !ok {
		return UpdateModeTrue, nil
	}
	switch t := v.(type) {
	case bool:
		if t {
			return UpdateModeTrue, nil
		}
		return UpdateModeFalse, nil
	case string:
		switch t {
		case "true":
			return UpdateModeTrue, nil
		case "false":
			return UpdateModeFalse, nil
		case "lazy":
			return UpdateModeLazy, nil
		}
	}
	return "", &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("invalid value for 'update': %v", v)}
}

// Reduce returns the reduce option.
func (o Map) Reduce() (*bool, error) {
	if group, _ := o.Group(); group {
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

// Group returns the group option.
func (o Map) Group() (bool, error) {
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

func (o Map) groupLevel() (uint64, error) {
	raw, ok := o["group_level"]
	if !ok {
		return 0, nil
	}
	return toUint64(raw, "invalid value for 'group_level'")
}

// Sort returns the sort option.
func (o Map) Sort() ([]string, error) {
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

// Bookmark returns the bookmark option.
func (o Map) Bookmark() (string, error) {
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

// Conflicts returns the conflicts option.
func (o Map) Conflicts() (bool, error) {
	if o.Meta() {
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

// Meta returns the meta option.
func (o Map) Meta() bool {
	v, _ := toBool(o["meta"])
	return v
}

// DeletedConflicts returns the deleted_conflicts option.
func (o Map) DeletedConflicts() bool {
	if o.Meta() {
		return true
	}
	v, _ := toBool(o["deleted_conflicts"])
	return v
}

// RevsInfo returns the revs_info option.
func (o Map) RevsInfo() bool {
	if o.Meta() {
		return true
	}
	v, _ := toBool(o["revs_info"])
	return v
}

// LocalSeq returns the local_seq option.
func (o Map) LocalSeq() bool {
	v, _ := toBool(o["local_seq"])
	return v
}

// AttsSince returns the atts_since option.
func (o Map) AttsSince() []string {
	attsSince, _ := o["atts_since"].([]string)
	return attsSince
}

// UpdateSeq returns the update_seq option.
func (o Map) UpdateSeq() (bool, error) {
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

// AttEncodingInfo returns the att_encoding_info option.
func (o Map) AttEncodingInfo() (bool, error) {
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

// BuildReduceCacheWhere returns WHERE conditions for use when querying the
// reduce cache.
func (v ViewOptions) BuildReduceCacheWhere(args *[]any) []string {
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

// BuildGroupWhere returns WHERE conditions for use with grouping.
func (v ViewOptions) BuildGroupWhere(args *[]any) []string {
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

// BookmarkWhere returns a WHERE clause for the bookmark option.
func (v ViewOptions) BookmarkWhere() string {
	if v.bookmark != "" {
		return `WHERE main.doc_number > (SELECT doc_number FROM bookmark)`
	}
	return ""
}

// BuildWhere returns WHERE conditions based on the provided configuration
// arguments, and may append to args as needed.
func (v ViewOptions) BuildWhere(args *[]any) []string {
	where := make([]string, 0, defaultWhereCap)
	switch v.view {
	case ViewAllDocs:
		where = append(where, `view.key NOT LIKE '"_local/%'`)
	case ViewLocalDocs:
		where = append(where, `view.key LIKE '"_local/%'`)
	case ViewDesignDocs:
		where = append(where, `view.key LIKE '"_design/%'`)
	}
	where = append(where, v.PaginationOptions.BuildWhere(args)...)
	if v.endkey != "" && v.endkeyDocID != "" {
		op := endKeyOp(v.descending, v.inclusiveEnd)
		where = append(where, fmt.Sprintf("view.id %s $%d", op, len(*args)+1))
		*args = append(*args, v.endkeyDocID)
	}
	if v.startkey != "" && v.startkeyDocID != "" {
		op := startKeyOp(v.descending)
		where = append(where, fmt.Sprintf("view.id %s $%d", op, len(*args)+1))
		*args = append(*args, v.startkeyDocID)
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

// ReduceGroupLevel returns the calculated groupLevel value to pass to
// [github.com/go-kivik/kivik/v4/x/sqlite/reduce.Reduce].
//
//	-1: Maximum grouping, same as group=true
//	 0: No grouping, same as group=false
//	1+: Group by the first N elements of the key, same as group_level=N
func (v ViewOptions) ReduceGroupLevel() int {
	if v.groupLevel == 0 && v.group {
		return -1
	}
	return int(v.groupLevel)
}

// ViewOptions are all of the options recognized by the view endpoints
// _design/<ddoc>/_view/<view>, _all_docs, _design_docs, and _local_docs.
//
// See https://docs.couchdb.org/en/stable/api/ddoc/views.html#api-ddoc-view
type ViewOptions struct {
	PaginationOptions
	view            string
	includeDocs     bool
	conflicts       bool
	reduce          *bool
	group           bool
	groupLevel      uint64
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

// FindOptions converts a _find query body into a viewOptions struct.
func FindOptions(query any) (*ViewOptions, error) {
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
	var o Map
	if err := json.Unmarshal(input, &o); err != nil {
		return nil, &internal.Error{Status: http.StatusBadRequest, Err: err}
	}

	limit, err := o.Limit()
	if err != nil {
		return nil, err
	}
	skip, err := o.Skip()
	if err != nil {
		return nil, err
	}
	fields, err := o.Fields()
	if err != nil {
		return nil, err
	}
	conflicts, err := o.Conflicts()
	if err != nil {
		return nil, err
	}
	bookmark, err := o.Bookmark()
	if err != nil {
		return nil, err
	}
	if bookmark != "" {
		skip = 0
	}
	sort, err := o.Sort()
	if err != nil {
		return nil, err
	}
	if len(sort) > 0 {
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: "no index exists for this sort, try indexing by the sort fields"}
	}

	v := &ViewOptions{
		PaginationOptions: PaginationOptions{
			limit: -1,
		},
		view:        ViewAllDocs,
		conflicts:   conflicts,
		includeDocs: true,
		findLimit:   limit,
		findSkip:    skip,
		selector:    s.Selector,
		fields:      fields,
		bookmark:    bookmark,
		sort:        sort,
	}

	return v, v.Validate()
}

// ViewOptions returns the viewOptions for the given view name.
func (o Map) ViewOptions(view string) (*ViewOptions, error) {
	pagination, err := o.PaginationOptions(false)
	if err != nil {
		return nil, err
	}
	reduce, err := o.Reduce()
	if err != nil {
		return nil, err
	}
	group, err := o.Group()
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
	conflicts, err := o.Conflicts()
	if err != nil {
		return nil, err
	}
	includeDocs, err := o.IncludeDocs()
	if err != nil {
		return nil, err
	}
	attachments, err := o.Attachments()
	if err != nil {
		return nil, err
	}
	update, err := o.Update()
	if err != nil {
		return nil, err
	}
	updateSeq, err := o.UpdateSeq()
	if err != nil {
		return nil, err
	}
	endkeyDocID, err := o.EndKeyDocID()
	if err != nil {
		return nil, err
	}
	startkeyDocID, err := o.StartKeyDocID()
	if err != nil {
		return nil, err
	}
	key, err := o.Key()
	if err != nil {
		return nil, err
	}
	keys, err := o.Keys()
	if err != nil {
		return nil, err
	}
	if len(keys) > 0 && (key != "" || pagination.endkey != "" || pagination.startkey != "") {
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: "`keys` is incompatible with `key`, `start_key` and `end_key`"}
	}
	sorted, err := o.Sorted()
	if err != nil {
		return nil, err
	}
	attEncodingInfo, err := o.AttEncodingInfo()
	if err != nil {
		return nil, err
	}

	v := &ViewOptions{
		PaginationOptions: *pagination,
		view:              view,
		includeDocs:       includeDocs,
		conflicts:         conflicts,
		reduce:            reduce,
		group:             group,
		groupLevel:        groupLevel,
		attachments:       attachments,
		update:            update,
		updateSeq:         updateSeq,
		endkeyDocID:       endkeyDocID,
		startkeyDocID:     startkeyDocID,
		key:               key,
		keys:              keys,
		sorted:            sorted,
		attEncodingInfo:   attEncodingInfo,
	}
	return v, v.Validate()
}

// Validate returns an error if the options are invalid.
func (v ViewOptions) Validate() error {
	if err := v.PaginationOptions.Validate(); err != nil {
		return err
	}
	descendingModifier := v.descendingModifier()

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

// IncludeDocs returns the include_docs option.
func (v ViewOptions) IncludeDocs() bool { return v.includeDocs }

// Conflicts returns the conflicts option.
func (v ViewOptions) Conflicts() bool { return v.conflicts }

// Attachments returns the attachments option.
func (v ViewOptions) Attachments() bool { return v.attachments }

// Update returns the update option.
func (v ViewOptions) Update() string { return v.update }

// UpdateSeq returns the update_seq option.
func (v ViewOptions) UpdateSeq() bool { return v.updateSeq }

// Reduce returns the reduce option.
func (v ViewOptions) Reduce() *bool { return v.reduce }

// Group returns the group option.
func (v ViewOptions) Group() bool { return v.group }

// GroupLevel returns the group_level option.
func (v ViewOptions) GroupLevel() uint64 { return v.groupLevel }

// Selector returns the selector option for _find queries.
func (v ViewOptions) Selector() *mango.Selector { return v.selector }

// Fields returns the fields option for _find queries.
func (v ViewOptions) Fields() []string { return v.fields }

// FindLimit returns the limit option for _find queries.
func (v ViewOptions) FindLimit() int64 { return v.findLimit }

// FindSkip returns the skip option for _find queries.
func (v ViewOptions) FindSkip() int64 { return v.findSkip }

// Bookmark returns the bookmark option for _find queries.
func (v ViewOptions) Bookmark() string { return v.bookmark }

// BuildOrderBy returns an ORDER BY clause based on the provided configuration.
// Additional columns can be specified to include in the ORDER BY clause.
func (v ViewOptions) BuildOrderBy(moreColumns ...string) string {
	if !v.sorted {
		return ""
	}
	direction := "ASC"
	if v.descending {
		direction = "DESC"
	}
	conditions := make([]string, 0, len(moreColumns)+1)
	for _, col := range append([]string{"key"}, moreColumns...) {
		conditions = append(conditions, "view."+col+" "+direction)
	}
	return "ORDER BY " + strings.Join(conditions, ", ")
}

func endKeyOp(descending, inclusive bool) string {
	switch {
	case descending && inclusive:
		return ">="
	case descending && !inclusive:
		return ">"
	case !descending && inclusive:
		return "<="
	case !descending && !inclusive:
		return "<"
	}
	panic("unreachable")
}

func startKeyOp(descending bool) string {
	if descending {
		return "<="
	}
	return ">="
}

func isBuiltinView(view string) bool {
	switch view {
	case ViewAllDocs, ViewLocalDocs, ViewDesignDocs:
		return true
	}
	return false
}

func couchdbCmpString(a, b string) int {
	return collate.CompareJSON(json.RawMessage(a), json.RawMessage(b))
}

func placeholders(start, count int) string {
	result := make([]string, count)
	for i := range result {
		result[i] = fmt.Sprintf("$%d", start+i)
	}
	return strings.Join(result, ", ")
}
