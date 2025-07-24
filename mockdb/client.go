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
	"errors"
	"time"

	"github.com/go-kivik/kivik/v4/driver"
)

type driverClient struct {
	*Client
}

var (
	_ driver.Client        = &driverClient{}
	_ driver.ClientCloser  = &driverClient{}
	_ driver.Cluster       = &driverClient{}
	_ driver.DBsStatser    = &driverClient{}
	_ driver.Pinger        = &driverClient{}
	_ driver.Sessioner     = &driverClient{}
	_ driver.Configer      = &driverClient{}
	_ driver.AllDBsStatser = &driverClient{}
)

func (c *driverClient) CreateDB(ctx context.Context, name string, options driver.Options) error {
	expected := &ExpectedCreateDB{
		arg0: name,
		commonExpectation: commonExpectation{
			options: options,
		},
	}
	if err := c.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback(ctx, name, options)
	}
	return expected.wait(ctx)
}

type driverReplication struct {
	*Replication
}

var _ driver.Replication = &driverReplication{}

func (r *driverReplication) ReplicationID() string {
	return r.id
}

func (r *driverReplication) Source() string {
	return r.source
}

func (r *driverReplication) Target() string {
	return r.target
}

func (r *driverReplication) StartTime() time.Time {
	return r.startTime
}

func (r *driverReplication) EndTime() time.Time {
	return r.endTime
}

func (r *driverReplication) State() string {
	return r.state
}

func (r *driverReplication) Err() error {
	return r.err
}

func (r *driverReplication) Delete(context.Context) error {
	return errors.New("not implemented")
}

func (r *driverReplication) Update(context.Context, *driver.ReplicationInfo) error {
	return errors.New("not implemented")
}

func driverReplications(in []*Replication) []driver.Replication {
	out := make([]driver.Replication, len(in))
	for i, r := range in {
		out[i] = &driverReplication{r}
	}
	return out
}
