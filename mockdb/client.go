package kivikmock

import (
	"context"
	"errors"
	"reflect"
	"time"

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

func (c *driverClient) CreateDB(ctx context.Context, name string, options map[string]interface{}) error {
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
	return r.Replication.id
}

func (r *driverReplication) Source() string {
	return r.Replication.source
}

func (r *driverReplication) Target() string {
	return r.Replication.target
}

func (r *driverReplication) StartTime() time.Time {
	return r.Replication.startTime
}

func (r *driverReplication) EndTime() time.Time {
	return r.Replication.endTime
}

func (r *driverReplication) State() string {
	return r.Replication.state
}

func (r *driverReplication) Err() error {
	return r.Replication.err
}

func (r *driverReplication) Delete(_ context.Context) error {
	return errors.New("not implemented")
}

func (r *driverReplication) Update(_ context.Context, _ *driver.ReplicationInfo) error {
	return errors.New("not implemented")
}

func driverReplications(in []*Replication) []driver.Replication {
	out := make([]driver.Replication, len(in))
	for i, r := range in {
		out[i] = &driverReplication{r}
	}
	return out
}
