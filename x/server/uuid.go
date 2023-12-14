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

package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.com/flimzy/httpe"

	"github.com/go-kivik/kivik/v4/internal"
)

type uuidAlgorithm string

const (
	// Supported UUID algorithms
	uuidAlgorithmRandom     uuidAlgorithm = "random"
	uuidAlgorithmSequential uuidAlgorithm = "sequential"
	uuidAlgorithmUTCRandom  uuidAlgorithm = "utc_random"
	uuidAlgorithmUTCID      uuidAlgorithm = "utc_id"
	uuidAlgorithmDefault                  = uuidAlgorithmSequential

	uuidDefaultMaxCount = 1000
)

var supportedUUIDAlgorithms = []uuidAlgorithm{uuidAlgorithmRandom, uuidAlgorithmSequential, uuidAlgorithmUTCRandom, uuidAlgorithmUTCID}

// confUUIDAlgorithm returns the UUID algorithm used by the server.
func (s *Server) confUUIDAlgorithm(ctx context.Context) uuidAlgorithm {
	var algo uuidAlgorithm
	_ = s.conf(ctx, "couchdb", "algorithm", &algo)
	for _, a := range supportedUUIDAlgorithms {
		if algo == a {
			return algo
		}
	}

	return uuidAlgorithmDefault
}

func (s *Server) confUUIDMaxCount(ctx context.Context) int {
	var count int
	_ = s.conf(ctx, "couchdb", "max_count", &count)
	if count < 1 {
		return uuidDefaultMaxCount
	}
	return count
}

func (s *Server) uuids() httpe.HandlerWithError {
	return httpe.HandlerWithErrorFunc(func(w http.ResponseWriter, r *http.Request) error {
		var count int
		if param := r.URL.Query().Get("count"); param != "" {
			var err error
			count, err = strconv.Atoi(param)
			if err != nil {
				return &internal.Error{Status: http.StatusBadRequest, Message: "count must be a positive integer"}
			}
		}
		maxCount := s.confUUIDMaxCount(r.Context())
		if count > maxCount {
			return &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("count must not exceed %d", maxCount)}
		}
		var uuids []string
		var err error
		switch algo := s.confUUIDAlgorithm(r.Context()); algo {
		case uuidAlgorithmRandom:
			uuids, err = s.uuidsRandom(r.Context())
		case uuidAlgorithmSequential:
			uuids, err = s.uuidsRandom(r.Context())
		case uuidAlgorithmUTCRandom:
			uuids, err = s.uuidsRandom(r.Context())
		case uuidAlgorithmUTCID:
			uuids, err = s.uuidsRandom(r.Context())
		}
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, map[string][]string{"uuids": uuids})
	})
}

func (s *Server) uuidsRandom(ctx context.Context) ([]string, error) {
	return []string{"x"}, nil
}
