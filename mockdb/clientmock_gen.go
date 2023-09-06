/* This file is auto-generated. Do not edit it! */

package kivikmock

import (
	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

var _ = kivik.EndKeySuffix // To ensure a reference to kivik package
var _ = &driver.Attachment{}

// ExpectAllDBs queues an expectation that AllDBs will be called.
func (c *Client) ExpectAllDBs() *ExpectedAllDBs {
	e := &ExpectedAllDBs{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectClose queues an expectation that Close will be called.
func (c *Client) ExpectClose() *ExpectedClose {
	e := &ExpectedClose{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectClusterSetup queues an expectation that ClusterSetup will be called.
func (c *Client) ExpectClusterSetup() *ExpectedClusterSetup {
	e := &ExpectedClusterSetup{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectClusterStatus queues an expectation that ClusterStatus will be called.
func (c *Client) ExpectClusterStatus() *ExpectedClusterStatus {
	e := &ExpectedClusterStatus{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectConfigValue queues an expectation that ConfigValue will be called.
func (c *Client) ExpectConfigValue() *ExpectedConfigValue {
	e := &ExpectedConfigValue{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectDBExists queues an expectation that DBExists will be called.
func (c *Client) ExpectDBExists() *ExpectedDBExists {
	e := &ExpectedDBExists{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectDeleteConfigKey queues an expectation that DeleteConfigKey will be called.
func (c *Client) ExpectDeleteConfigKey() *ExpectedDeleteConfigKey {
	e := &ExpectedDeleteConfigKey{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectDestroyDB queues an expectation that DestroyDB will be called.
func (c *Client) ExpectDestroyDB() *ExpectedDestroyDB {
	e := &ExpectedDestroyDB{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectPing queues an expectation that Ping will be called.
func (c *Client) ExpectPing() *ExpectedPing {
	e := &ExpectedPing{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectSetConfigValue queues an expectation that SetConfigValue will be called.
func (c *Client) ExpectSetConfigValue() *ExpectedSetConfigValue {
	e := &ExpectedSetConfigValue{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectConfig queues an expectation that Config will be called.
func (c *Client) ExpectConfig() *ExpectedConfig {
	e := &ExpectedConfig{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectConfigSection queues an expectation that ConfigSection will be called.
func (c *Client) ExpectConfigSection() *ExpectedConfigSection {
	e := &ExpectedConfigSection{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectDB queues an expectation that DB will be called.
func (c *Client) ExpectDB() *ExpectedDB {
	e := &ExpectedDB{
		ret0: &DB{},
	}
	c.expected = append(c.expected, e)
	return e
}

// ExpectDBUpdates queues an expectation that DBUpdates will be called.
func (c *Client) ExpectDBUpdates() *ExpectedDBUpdates {
	e := &ExpectedDBUpdates{
		ret0: &Updates{},
	}
	c.expected = append(c.expected, e)
	return e
}

// ExpectDBsStats queues an expectation that DBsStats will be called.
func (c *Client) ExpectDBsStats() *ExpectedDBsStats {
	e := &ExpectedDBsStats{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectGetReplications queues an expectation that GetReplications will be called.
func (c *Client) ExpectGetReplications() *ExpectedGetReplications {
	e := &ExpectedGetReplications{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectMembership queues an expectation that Membership will be called.
func (c *Client) ExpectMembership() *ExpectedMembership {
	e := &ExpectedMembership{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectReplicate queues an expectation that Replicate will be called.
func (c *Client) ExpectReplicate() *ExpectedReplicate {
	e := &ExpectedReplicate{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectSession queues an expectation that Session will be called.
func (c *Client) ExpectSession() *ExpectedSession {
	e := &ExpectedSession{}
	c.expected = append(c.expected, e)
	return e
}

// ExpectVersion queues an expectation that Version will be called.
func (c *Client) ExpectVersion() *ExpectedVersion {
	e := &ExpectedVersion{}
	c.expected = append(c.expected, e)
	return e
}
