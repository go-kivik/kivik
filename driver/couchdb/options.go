package couchdb

import (
	"errors"
	"fmt"
)

// Available options
const (
	optionForceCommit = "force_commit"
)

// SetDefault allows setting default database options.
func (c *client) SetDefault(key string, value interface{}) error {
	switch key {
	case optionForceCommit:
		return c.setForceCommit(value)
	}
	return errors.New("unknown option")
}

func (c *client) setForceCommit(value interface{}) error {
	if force, ok := value.(bool); ok {
		c.forceCommit = force
		return nil
	}
	return fmt.Errorf("invalid type %t for option %s", value, optionForceCommit)
}

func (d *db) SetOption(key string, value interface{}) error {
	switch key {
	case optionForceCommit:
		return d.setForceCommit(value)
	}
	return errors.New("unknown option")
}

func (d *db) setForceCommit(value interface{}) error {
	if force, ok := value.(bool); ok {
		d.forceCommit = force
		return nil
	}
	return fmt.Errorf("invalid type %t for options %s", value, optionForceCommit)
}
