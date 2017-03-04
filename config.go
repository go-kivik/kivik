package kivik

import "github.com/flimzy/kivik/driver"

// Config represents the entire server config
type Config map[string]ConfigSection

// ConfigSection represents a section of config
type ConfigSection map[string]string

// GetAllConfig returns the entire server config
func (c *Client) GetAllConfig() (Config, error) {
	if conf, ok := c.driverClient.(driver.Configer); ok {
		c, err := conf.GetAllConfig()
		if err != nil {
			return nil, err
		}
		cf := Config{}
		for section, config := range c {
			cf[section] = config
		}
		return cf, nil
	}
	return nil, ErrNotImplemented
}
