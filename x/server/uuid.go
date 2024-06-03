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

//go:build !js
// +build !js

package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"gitlab.com/flimzy/httpe"

	internal "github.com/go-kivik/kivik/v4/int/errors"
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
	value, _ := s.config.Key(ctx, "uuids", "algorithm")
	algo := uuidAlgorithm(value)
	for _, a := range supportedUUIDAlgorithms {
		if algo == a {
			return algo
		}
	}

	return uuidAlgorithmDefault
}

func (s *Server) confUUIDMaxCount(ctx context.Context) int {
	var count int
	_ = s.conf(ctx, "uuids", "max_count", &count)
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
		if count == 0 {
			count = 1
		}
		maxCount := s.confUUIDMaxCount(r.Context())
		if count > maxCount {
			return &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("count must not exceed %d", maxCount)}
		}
		var uuids []string
		var err error
		switch algo := s.confUUIDAlgorithm(r.Context()); algo {
		case uuidAlgorithmRandom:
			uuids, err = s.uuidsRandom(count)
		case uuidAlgorithmSequential:
			uuids, err = s.uuidsSequential(count)
		case uuidAlgorithmUTCRandom:
			uuids, err = s.uuidsUTCRandom(count)
		case uuidAlgorithmUTCID:
			uuids = s.uuidsUTCID(r.Context(), count)
		}
		if err != nil {
			return err
		}
		return serveJSON(w, http.StatusOK, map[string][]string{"uuids": uuids})
	})
}

func (s *Server) uuidsRandom(count int) ([]string, error) {
	const randomUUIDLength = 32
	uuids := make([]string, count)
	for i := range uuids {
		uuid, err := randomHexString(randomUUIDLength)
		if err != nil {
			return nil, err
		}
		uuids[i] = uuid
	}
	return uuids, nil
}

func randomHexString(length int) (string, error) {
	const charsPerByte = 2
	var hexByteLength int
	if length%4 == 0 {
		hexByteLength = length / charsPerByte
	} else {
		hexByteLength = length/charsPerByte + 1
	}
	randomBytes := make([]byte, hexByteLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes)[:length], nil
}

// randomIncrement returns a random number between 1 and 256.
func randomIncrement() (int32, error) {
	buf := make([]byte, 1)
	if _, err := rand.Read(buf); err != nil {
		return 0, err
	}
	return int32(buf[0]) + 1, nil
}

func (s *Server) uuidsSequential(count int) ([]string, error) {
	prefix, err := s.uuidSequentialPrefix()
	if err != nil {
		return nil, err
	}

	uuids := make([]string, count)
	for i := range uuids {
		incr, err := randomIncrement()
		if err != nil {
			return nil, err
		}
		monotonic := atomic.AddInt32(&s.sequentialUUIDMonotonicID, incr)
		uuids[i] = prefix + fmt.Sprintf("%06x", monotonic)
	}
	return uuids, nil
}

func (s *Server) uuidSequentialPrefix() (string, error) {
	s.sequentialMU.Lock()
	defer s.sequentialMU.Unlock()
	if s.sequentialUUIDPrefix != "" {
		return s.sequentialUUIDPrefix, nil
	}

	const sequentialUUIDPrefixLength = 26
	var err error
	s.sequentialUUIDPrefix, err = randomHexString(sequentialUUIDPrefixLength)
	return s.sequentialUUIDPrefix, err
}

func (s *Server) uuidsUTCRandom(count int) ([]string, error) {
	const randomChars = 18
	uuids := make([]string, count)
	for i := range uuids {
		r, err := randomHexString(randomChars)
		if err != nil {
			return nil, err
		}
		uuids[i] = fmt.Sprintf("%014x%s", time.Now().UnixMicro(), r)
	}
	return uuids, nil
}

func (s *Server) uuidsUTCID(ctx context.Context, count int) []string {
	suffix, _ := s.config.Key(ctx, "uuids", "utc_id_suffix")
	uuids := make([]string, count)
	for i := range uuids {
		uuids[i] = fmt.Sprintf("%014x%s", time.Now().UnixMicro(), suffix)
	}
	return uuids
}
