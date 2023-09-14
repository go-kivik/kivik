package mockdb

import (
	"context"
	"errors"

	"github.com/go-kivik/kivik/v4/driver"
)

type driverClient struct {
	*Client
}

var (
	_ driver.Client       = &driverClient{}
	_ driver.ClientCloser = &driverClient{}
	_ driver.Cluster      = &driverClient{}
	_ driver.DBsStatser   = &driverClient{}
	_ driver.Pinger       = &driverClient{}
	_ driver.Sessioner    = &driverClient{}
	_ driver.Configer     = &driverClient{}
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

func (r *driverReplication) Metadata() driver.ReplicationMetadata {
	return r.Replication.meta
}

func (r *driverReplication) State() string {
	return r.Replication.state
}

func (r *driverReplication) Err() error {
	return r.Replication.err
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
