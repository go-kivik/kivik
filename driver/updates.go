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

package driver

import "context"

// DBUpdate represents a database update event.
type DBUpdate struct {
	DBName string `json:"db_name"`
	Type   string `json:"type"`
	Seq    string `json:"seq"`
}

// DBUpdates is a DBUpdates iterator.
type DBUpdates interface {
	// Next is called to populate DBUpdate with the values of the next update in
	// the feed.
	//
	// Next should return [io.EOF] when the feed is closed normally.
	Next(*DBUpdate) error
	// Close closes the iterator.
	Close() error
}

// LastSeqer extends the [DBUpdates] interface, and in Kivik v5, will be
// included in it.
type LastSeqer interface {
	// LastSeq returns the last sequence ID reported.
	LastSeq() (string, error)
}

// DBUpdater is an optional interface that may be implemented by a
// [Client] to provide access to the DB Updates feed.
type DBUpdater interface {
	// DBUpdates must return a [DBUpdates] iterator. The context, or the iterator's
	// Close method, may be used to close the iterator.
	DBUpdates(ctx context.Context, options Options) (DBUpdates, error)
}
