package mock

import (
	"context"

	"github.com/go-kivik/kivik/driver"
)

// Client mocks driver.Client
type Client struct {
	// ID identifies a specific Client instance
	ID            string
	AllDBsFunc    func(context.Context, map[string]interface{}) ([]string, error)
	CreateDBFunc  func(context.Context, string, map[string]interface{}) error
	DBFunc        func(context.Context, string, map[string]interface{}) (driver.DB, error)
	DBExistsFunc  func(context.Context, string, map[string]interface{}) (bool, error)
	DestroyDBFunc func(context.Context, string, map[string]interface{}) error
	VersionFunc   func(context.Context) (*driver.Version, error)
}

var _ driver.Client = &Client{}

// AllDBs calls c.AllDBsFunc
func (c *Client) AllDBs(ctx context.Context, opts map[string]interface{}) ([]string, error) {
	return c.AllDBsFunc(ctx, opts)
}

// CreateDB calls c.CreateDBFunc
func (c *Client) CreateDB(ctx context.Context, dbname string, opts map[string]interface{}) error {
	return c.CreateDBFunc(ctx, dbname, opts)
}

// DB calls c.DBFunc
func (c *Client) DB(ctx context.Context, dbname string, opts map[string]interface{}) (driver.DB, error) {
	return c.DBFunc(ctx, dbname, opts)
}

// DBExists calls c.DBExistsFunc
func (c *Client) DBExists(ctx context.Context, dbname string, opts map[string]interface{}) (bool, error) {
	return c.DBExistsFunc(ctx, dbname, opts)
}

// DestroyDB calls c.DestroyDBFunc
func (c *Client) DestroyDB(ctx context.Context, dbname string, opts map[string]interface{}) error {
	return c.DestroyDBFunc(ctx, dbname, opts)
}

// Version calls c.VersionFunc
func (c *Client) Version(ctx context.Context) (*driver.Version, error) {
	return c.VersionFunc(ctx)
}

// ClientReplicator mocks driver.Client and driver.ClientReplicator
type ClientReplicator struct {
	*Client
	GetReplicationsFunc func(context.Context, map[string]interface{}) ([]driver.Replication, error)
	ReplicateFunc       func(context.Context, string, string, map[string]interface{}) (driver.Replication, error)
}

var _ driver.ClientReplicator = &ClientReplicator{}

// GetReplications calls c.GetReplicationsFunc
func (c *ClientReplicator) GetReplications(ctx context.Context, opts map[string]interface{}) ([]driver.Replication, error) {
	return c.GetReplicationsFunc(ctx, opts)
}

// Replicate calls c.ReplicateFunc
func (c *ClientReplicator) Replicate(ctx context.Context, target, source string, opts map[string]interface{}) (driver.Replication, error) {
	return c.ReplicateFunc(ctx, target, source, opts)
}

// Authenticator mocks driver.Client and driver.Authenticator
type Authenticator struct {
	*Client
	AuthenticateFunc func(context.Context, interface{}) error
}

var _ driver.Authenticator = &Authenticator{}

// Authenticate calls c.AuthenticateFunc
func (c *Authenticator) Authenticate(ctx context.Context, a interface{}) error {
	return c.AuthenticateFunc(ctx, a)
}

// DBUpdater mocks driver.Client and driver.DBUpdater
type DBUpdater struct {
	*Client
	DBUpdatesFunc func(context.Context) (driver.DBUpdates, error)
}

var _ driver.DBUpdater = &DBUpdater{}

// DBUpdates calls c.DBUpdatesFunc
func (c *DBUpdater) DBUpdates(ctx context.Context) (driver.DBUpdates, error) {
	return c.DBUpdatesFunc(ctx)
}

// DBsStatser mocks driver.Client and driver.DBsStatser
type DBsStatser struct {
	*Client
	DBsStatsFunc func(context.Context, []string) ([]*driver.DBStats, error)
}

var _ driver.DBsStatser = &DBsStatser{}

// DBsStats calls c.DBsStatsFunc
func (c *DBsStatser) DBsStats(ctx context.Context, dbnames []string) ([]*driver.DBStats, error) {
	return c.DBsStatsFunc(ctx, dbnames)
}

// Pinger mocks driver.Client and driver.Pinger
type Pinger struct {
	*Client
	PingFunc func(context.Context) (bool, error)
}

var _ driver.Pinger = &Pinger{}

// Ping calls c.PingFunc
func (c *Pinger) Ping(ctx context.Context) (bool, error) {
	return c.PingFunc(ctx)
}

// Cluster mocks driver.Client and driver.Cluster
type Cluster struct {
	*Client
	ClusterStatusFunc func(context.Context, map[string]interface{}) (string, error)
	ClusterSetupFunc  func(context.Context, interface{}) error
}

var _ driver.Cluster = &Cluster{}

// ClusterStatus calls c.ClusterStatusFunc
func (c *Cluster) ClusterStatus(ctx context.Context, options map[string]interface{}) (string, error) {
	return c.ClusterStatusFunc(ctx, options)
}

// ClusterSetup calls c.ClusterSetupFunc
func (c *Cluster) ClusterSetup(ctx context.Context, action interface{}) error {
	return c.ClusterSetupFunc(ctx, action)
}

// ClientCloser mocks driver.Client and driver.ClientCloser
type ClientCloser struct {
	*Client
	CloseFunc func(context.Context) error
}

var _ driver.ClientCloser = &ClientCloser{}

// Close calls c.CloseFunc
func (c *ClientCloser) Close(ctx context.Context) error {
	return c.CloseFunc(ctx)
}

type Configer struct {
	*Client
	ConfigFunc          func(context.Context, string) (driver.Config, error)
	ConfigSectionFunc   func(context.Context, string, string) (driver.ConfigSection, error)
	ConfigValueFunc     func(context.Context, string, string, string) (string, error)
	SetConfigValueFunc  func(context.Context, string, string, string, string) (string, error)
	DeleteConfigKeyFunc func(context.Context, string, string, string) (string, error)
}

var _ driver.Configer = &Configer{}

// Config calls c.ConfigFunc
func (c *Configer) Config(ctx context.Context, node string) (driver.Config, error) {
	return c.ConfigFunc(ctx, node)
}

// ConfigSection calls c.ConfSectionFunc
func (c *Configer) ConfigSection(ctx context.Context, node, section string) (driver.ConfigSection, error) {
	return c.ConfigSectionFunc(ctx, node, section)
}

// ConfigValue calls c.ConfigValueFunc
func (c *Configer) ConfigValue(ctx context.Context, node, section, key string) (string, error) {
	return c.ConfigValueFunc(ctx, node, section, key)
}

// SetConfigValue calls c.SetConfigValueFunc
func (c *Configer) SetConfigValue(ctx context.Context, node, section, key, value string) (string, error) {
	return c.SetConfigValueFunc(ctx, node, section, key, value)
}

// DeleteConfigKey calls c.DeleteConfigKeyFunc
func (c *Configer) DeleteConfigKey(ctx context.Context, node, section, key string) (string, error) {
	return c.DeleteConfigKeyFunc(ctx, node, section, key)
}
