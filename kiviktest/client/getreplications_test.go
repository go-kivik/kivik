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

package client

import (
	"sync"
	"testing"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func TestLockReplicationRace(t *testing.T) {
	t.Cleanup(func() {
		replicationMUs = make(map[*kivik.Client]*sync.Mutex)
	})
	replicationMUs = make(map[*kivik.Client]*sync.Mutex)

	clients := make([]*kivik.Client, 10)
	for i := range clients {
		clients[i] = &kivik.Client{}
	}

	var wg sync.WaitGroup
	for _, c := range clients {
		c := c
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := &kt.Context{ContextCore: &kt.ContextCore{Admin: c}}
			unlock := lockReplication(ctx)
			unlock()
		}()
	}
	wg.Wait()
}
