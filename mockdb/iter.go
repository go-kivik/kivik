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

package mockdb

import (
	"context"
	"io"
	"time"
)

type item struct {
	delay time.Duration
	item  interface{}
}

type iter struct {
	items     []*item
	closeErr  error
	resultErr error
}

func (i *iter) Close() error { return i.closeErr }

func (i *iter) push(item *item) {
	i.items = append(i.items, item)
}

func (i *iter) unshift(ctx context.Context) (interface{}, error) {
	if len(i.items) == 0 {
		if i.resultErr != nil {
			return nil, i.resultErr
		}
		return nil, io.EOF
	}
	var item *item
	item, i.items = i.items[0], i.items[1:]
	if item.delay == 0 {
		return item.item, nil
	}
	if err := pause(ctx, item.delay); err != nil {
		return nil, err
	}
	return i.unshift(ctx)
}

func (i *iter) count() int {
	if len(i.items) == 0 {
		return 0
	}
	var count int
	for _, result := range i.items {
		if result != nil && result.item != nil {
			count++
		}
	}

	return count
}
