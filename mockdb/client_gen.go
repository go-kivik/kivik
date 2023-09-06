/* This file is auto-generated. Do not edit it! */

package kivikmock

import (
	"context"

	"github.com/go-kivik/kivik/v4/driver"
)

var _ = &driver.Attachment{}

func (c *driverClient) AllDBs(ctx context.Context, options map[string]interface{}) ([]string, error) {
	expected := &ExpectedAllDBs{
		commonExpectation: commonExpectation{
			options: options,
		},
	}
	if err := c.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (c *driverClient) Close() error {
	expected := &ExpectedClose{}
	if err := c.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback()
	}
	return expected.err
}

func (c *driverClient) ClusterSetup(ctx context.Context, arg0 interface{}) error {
	expected := &ExpectedClusterSetup{
		arg0: arg0,
	}
	if err := c.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0)
	}
	return expected.wait(ctx)
}

func (c *driverClient) ClusterStatus(ctx context.Context, options map[string]interface{}) (string, error) {
	expected := &ExpectedClusterStatus{
		commonExpectation: commonExpectation{
			options: options,
		},
	}
	if err := c.nextExpectation(expected); err != nil {
		return "", err
	}
	if expected.callback != nil {
		return expected.callback(ctx, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (c *driverClient) ConfigValue(ctx context.Context, arg0 string, arg1 string, arg2 string) (string, error) {
	expected := &ExpectedConfigValue{
		arg0: arg0,
		arg1: arg1,
		arg2: arg2,
	}
	if err := c.nextExpectation(expected); err != nil {
		return "", err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, arg2)
	}
	return expected.ret0, expected.wait(ctx)
}

func (c *driverClient) DBExists(ctx context.Context, arg0 string, options map[string]interface{}) (bool, error) {
	expected := &ExpectedDBExists{
		arg0: arg0,
		commonExpectation: commonExpectation{
			options: options,
		},
	}
	if err := c.nextExpectation(expected); err != nil {
		return false, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (c *driverClient) DeleteConfigKey(ctx context.Context, arg0 string, arg1 string, arg2 string) (string, error) {
	expected := &ExpectedDeleteConfigKey{
		arg0: arg0,
		arg1: arg1,
		arg2: arg2,
	}
	if err := c.nextExpectation(expected); err != nil {
		return "", err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, arg2)
	}
	return expected.ret0, expected.wait(ctx)
}

func (c *driverClient) DestroyDB(ctx context.Context, arg0 string, options map[string]interface{}) error {
	expected := &ExpectedDestroyDB{
		arg0: arg0,
		commonExpectation: commonExpectation{
			options: options,
		},
	}
	if err := c.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, options)
	}
	return expected.wait(ctx)
}

func (c *driverClient) Ping(ctx context.Context) (bool, error) {
	expected := &ExpectedPing{}
	if err := c.nextExpectation(expected); err != nil {
		return false, err
	}
	if expected.callback != nil {
		return expected.callback(ctx)
	}
	return expected.ret0, expected.wait(ctx)
}

func (c *driverClient) SetConfigValue(ctx context.Context, arg0 string, arg1 string, arg2 string, arg3 string) (string, error) {
	expected := &ExpectedSetConfigValue{
		arg0: arg0,
		arg1: arg1,
		arg2: arg2,
		arg3: arg3,
	}
	if err := c.nextExpectation(expected); err != nil {
		return "", err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, arg2, arg3)
	}
	return expected.ret0, expected.wait(ctx)
}

func (c *driverClient) Config(ctx context.Context, arg0 string) (driver.Config, error) {
	expected := &ExpectedConfig{
		arg0: arg0,
	}
	if err := c.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0)
	}
	return expected.ret0, expected.wait(ctx)
}

func (c *driverClient) ConfigSection(ctx context.Context, arg0 string, arg1 string) (driver.ConfigSection, error) {
	expected := &ExpectedConfigSection{
		arg0: arg0,
		arg1: arg1,
	}
	if err := c.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1)
	}
	return expected.ret0, expected.wait(ctx)
}

func (c *driverClient) DB(arg0 string, options map[string]interface{}) (driver.DB, error) {
	expected := &ExpectedDB{
		arg0: arg0,
		commonExpectation: commonExpectation{
			options: options,
		},
	}
	if err := c.nextExpectation(expected); err != nil {
		return nil, err
	}
	expected.ret0.mu.Lock()
	expected.ret0.name = arg0
	expected.ret0.mu.Unlock()
	if expected.callback != nil {
		return expected.callback(arg0, options)
	}
	return &driverDB{DB: expected.ret0}, expected.err
}

func (c *driverClient) DBUpdates(ctx context.Context, options map[string]interface{}) (driver.DBUpdates, error) {
	expected := &ExpectedDBUpdates{
		commonExpectation: commonExpectation{
			options: options,
		},
	}
	if err := c.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, options)
	}
	return &driverDBUpdates{Context: ctx, Updates: coalesceDBUpdates(expected.ret0)}, expected.wait(ctx)
}

func (c *driverClient) DBsStats(ctx context.Context, arg0 []string) ([]*driver.DBStats, error) {
	expected := &ExpectedDBsStats{
		arg0: arg0,
	}
	if err := c.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0)
	}
	return expected.ret0, expected.wait(ctx)
}

func (c *driverClient) GetReplications(ctx context.Context, options map[string]interface{}) ([]driver.Replication, error) {
	expected := &ExpectedGetReplications{
		commonExpectation: commonExpectation{
			options: options,
		},
	}
	if err := c.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, options)
	}
	return driverReplications(expected.ret0), expected.wait(ctx)
}

func (c *driverClient) Membership(ctx context.Context) (*driver.ClusterMembership, error) {
	expected := &ExpectedMembership{}
	if err := c.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx)
	}
	return expected.ret0, expected.wait(ctx)
}

func (c *driverClient) Replicate(ctx context.Context, arg0 string, arg1 string, options map[string]interface{}) (driver.Replication, error) {
	expected := &ExpectedReplicate{
		arg0: arg0,
		arg1: arg1,
		commonExpectation: commonExpectation{
			options: options,
		},
	}
	if err := c.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, options)
	}
	return &driverReplication{Replication: expected.ret0}, expected.wait(ctx)
}

func (c *driverClient) Session(ctx context.Context) (*driver.Session, error) {
	expected := &ExpectedSession{}
	if err := c.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx)
	}
	return expected.ret0, expected.wait(ctx)
}

func (c *driverClient) Version(ctx context.Context) (*driver.Version, error) {
	expected := &ExpectedVersion{}
	if err := c.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx)
	}
	return expected.ret0, expected.wait(ctx)
}
