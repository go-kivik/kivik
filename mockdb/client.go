package mockdb

import (
	"context"
	"errors"
	"reflect"

	"github.com/go-kivik/kivik/v4/driver"
)

type driverClient struct {
	*Client
}

var (
	_ driver.Client        = &driverClient{}
	_ driver.ClientCloser  = &driverClient{}
	_ driver.Authenticator = &driverClient{}
	_ driver.Cluster       = &driverClient{}
	_ driver.DBsStatser    = &driverClient{}
	_ driver.Pinger        = &driverClient{}
	_ driver.Sessioner     = &driverClient{}
	_ driver.Configer      = &driverClient{}
)

func (c *driverClient) Authenticate(ctx context.Context, authenticator interface{}) error {
	expected := &ExpectedAuthenticate{
		authType: reflect.TypeOf(authenticator).Name(),
	}
	if err := c.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback(ctx, authenticator)
	}
	return expected.wait(ctx)
}

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
	return driver.ReplicationMetadata{
		ID:        r.Replication.id,
		Source:    r.Replication.source,
		Target:    r.Replication.target,
		StartTime: r.Replication.startTime,
		EndTime:   r.Replication.endTime,
		State:     r.Replication.state,
		Error:     r.Replication.err,
	}
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
