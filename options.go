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

package kivik

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-kivik/kivik/v4/driver"
)

// Option is a Kivik or driver option.
//
// Most methods/endpoints take query parameters which are passed as part of
// the query URL, as documented in the official CouchDB documentation. You can
// use [Params] or [Param] to set arbitrary query parameters. Backend drivers
// may provide their own special-purpose options as well.
type Option interface {
	// Apply applies the option to target, if target is of the expected type.
	// Unexpected/recognized target types should be ignored.
	Apply(target interface{})
}

var _ Option = (driver.Options)(nil)

type multiOptions []Option

var _ Option = (multiOptions)(nil)

func (o multiOptions) Apply(t interface{}) {
	for _, opt := range o {
		if opt != nil {
			opt.Apply(t)
		}
	}
}

func (o multiOptions) String() string {
	parts := make([]string, 0, len(o))
	for _, opt := range o {
		if o != nil {
			if part := fmt.Sprintf("%s", opt); part != "" {
				parts = append(parts, part)
			}
		}
	}
	return strings.Join(parts, ",")
}

type params map[string]interface{}

// Apply applies o to target. The following target types are supported:
//
//   - map[string]interface{}
//   - *url.Values
func (o params) Apply(target interface{}) {
	switch t := target.(type) {
	case map[string]interface{}:
		for k, v := range o {
			t[k] = v
		}
	case *url.Values:
		for key, i := range o {
			var values []string
			switch v := i.(type) {
			case string:
				values = []string{v}
			case []string:
				values = v
			case bool:
				values = []string{fmt.Sprintf("%t", v)}
			case int, uint, uint8, uint16, uint32, uint64, int8, int16, int32, int64:
				values = []string{fmt.Sprintf("%d", v)}
			case float64:
				values = []string{strconv.FormatFloat(v, 'f', -1, 64)}
			case float32:
				values = []string{strconv.FormatFloat(float64(v), 'f', -1, 32)}
			default:
				panic(fmt.Sprintf("kivik: unknown option type: %T", v))
			}
			for _, value := range values {
				t.Add(key, value)
			}
		}
	}
}

func (o params) String() string {
	if len(o) == 0 {
		return ""
	}
	return fmt.Sprintf("%v", map[string]interface{}(o))
}

// Params allows passing a collection of key/value pairs as query parameter
// options.
func Params(p map[string]interface{}) Option {
	return params(p)
}

// Param sets a single key/value pair as a query parameter.
func Param(key string, value interface{}) Option {
	return params{key: value}
}

// Rev is a convenience function to set the revision. A less verbose alternative
// to Param("rev", rev).
func Rev(rev string) Option {
	return params{"rev": rev}
}

// IncludeDocs instructs the query to include documents. A less verbose
// alternative to Param("include_docs", true).
func IncludeDocs() Option {
	return params{"include_docs": true}
}

type durationParam struct {
	key   string
	value time.Duration
}

var _ Option = durationParam{}

// Apply supports map[string]interface{} and *url.Values targets.
func (p durationParam) Apply(target interface{}) {
	switch t := target.(type) {
	case map[string]interface{}:
		t[p.key] = fmt.Sprintf("%d", p.value/time.Millisecond)
	case *url.Values:
		t.Add(p.key, fmt.Sprintf("%d", p.value/time.Millisecond))
	}
}

func (p durationParam) String() string {
	return fmt.Sprintf("[%s=%s]", p.key, p.value)
}

// Duration is a convenience function for setting query parameters from
// [time.Duration] values. The duration will be converted to milliseconds when
// passed as a query parameter.
//
// For example, Duration("heartbeat", 15 * time.Second) will result in appending
// ?heartbeat=15000 to the query.
func Duration(key string, dur time.Duration) Option {
	return durationParam{
		key:   key,
		value: dur,
	}
}
