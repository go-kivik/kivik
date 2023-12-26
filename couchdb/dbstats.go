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

package couchdb

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivik/v4/driver"
)

type dbStats struct {
	driver.DBStats
	Sizes struct {
		File     int64 `json:"file"`
		External int64 `json:"external"`
		Active   int64 `json:"active"`
	} `json:"sizes"`
	UpdateSeq json.RawMessage `json:"update_seq"` // nolint: govet
	rawBody   json.RawMessage
}

func (s *dbStats) UnmarshalJSON(p []byte) error {
	type dbStatsClone dbStats
	c := dbStatsClone(*s)
	if err := json.Unmarshal(p, &c); err != nil {
		return err
	}
	*s = dbStats(c)
	s.rawBody = p
	return nil
}

func (s *dbStats) driverStats() *driver.DBStats {
	stats := &s.DBStats
	if s.Sizes.File > 0 {
		stats.DiskSize = s.Sizes.File
	}
	if s.Sizes.External > 0 {
		stats.ExternalSize = s.Sizes.External
	}
	if s.Sizes.Active > 0 {
		stats.ActiveSize = s.Sizes.Active
	}
	stats.UpdateSeq = string(bytes.Trim(s.UpdateSeq, `"`))
	stats.RawResponse = s.rawBody
	return stats
}

func (d *db) Stats(ctx context.Context) (*driver.DBStats, error) {
	result := dbStats{}
	if err := d.Client.DoJSON(ctx, http.MethodGet, d.dbName, nil, &result); err != nil {
		return nil, err
	}
	return result.driverStats(), nil
}

type dbsInfoRequest struct {
	Keys []string `json:"keys"`
}

type dbsInfoResponse struct {
	Key    string  `json:"key"`
	DBInfo dbStats `json:"info"`
	Error  string  `json:"error"`
}

func (c *client) DBsStats(_ context.Context, dbnames []string) ([]*driver.DBStats, error) {
	opts := &chttp.Options{
		GetBody: chttp.BodyEncoder(dbsInfoRequest{Keys: dbnames}),
		Header: http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		},
	}
	result := []dbsInfoResponse{}
	err := c.DoJSON(context.Background(), http.MethodPost, "/_dbs_info", opts, &result)
	if err != nil {
		return nil, err
	}
	stats := make([]*driver.DBStats, len(result))
	for i := range result {
		if result[i].Error == "" {
			stats[i] = result[i].DBInfo.driverStats()
		}
	}
	return stats, nil
}

func (c *client) AllDBsStats(ctx context.Context, options driver.Options) ([]*driver.DBStats, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)
	chttpOpts := &chttp.Options{
		Header: http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		},
	}
	var err error
	chttpOpts.Query, err = optionsToParams(opts)
	if err != nil {
		return nil, err
	}
	result := []dbsInfoResponse{}
	if err := c.DoJSON(ctx, http.MethodGet, "/_dbs_info", chttpOpts, &result); err != nil {
		return nil, err
	}
	stats := make([]*driver.DBStats, len(result))
	for i := range result {
		stats[i] = result[i].DBInfo.driverStats()
	}
	return stats, nil
}

type partitionStats struct {
	DBName      string `json:"db_name"`
	DocCount    int64  `json:"doc_count"`
	DocDelCount int64  `json:"doc_del_count"`
	Partition   string `json:"partition"`
	Sizes       struct {
		Active   int64 `json:"active"`
		External int64 `json:"external"`
	}
	rawBody json.RawMessage
}

func (s *partitionStats) UnmarshalJSON(p []byte) error {
	c := struct {
		partitionStats
		UnmarshalJSON struct{}
	}{}
	if err := json.Unmarshal(p, &c); err != nil {
		return err
	}
	*s = c.partitionStats
	s.rawBody = p
	return nil
}

func (d *db) PartitionStats(ctx context.Context, name string) (*driver.PartitionStats, error) {
	result := partitionStats{}
	if err := d.Client.DoJSON(ctx, http.MethodGet, d.path("_partition/"+name), nil, &result); err != nil {
		return nil, err
	}
	return &driver.PartitionStats{
		DBName:          result.DBName,
		DocCount:        result.DocCount,
		DeletedDocCount: result.DocDelCount,
		Partition:       result.Partition,
		ActiveSize:      result.Sizes.Active,
		ExternalSize:    result.Sizes.External,
		RawResponse:     result.rawBody,
	}, nil
}
