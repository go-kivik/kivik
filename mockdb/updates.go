package kivikmock

import (
	"context"
	"time"

	"github.com/go-kivik/kivik/v4/driver"
)

// Updates is a mocked collection of database updates.
type Updates struct {
	iter
}

type driverDBUpdates struct {
	context.Context
	*Updates
}

func coalesceDBUpdates(updates *Updates) *Updates {
	if updates != nil {
		return updates
	}
	return &Updates{}
}

var _ driver.DBUpdates = &driverDBUpdates{}

func (u *driverDBUpdates) Next(update *driver.DBUpdate) error {
	result, err := u.unshift(u.Context)
	if err != nil {
		return err
	}
	*update = *result.(*driver.DBUpdate)
	return nil
}

// CloseError sets an error to be returned when the updates iterator is closed.
func (u *Updates) CloseError(err error) *Updates {
	u.closeErr = err
	return u
}

// AddUpdateError adds an error to be returned during update iteration.
func (u *Updates) AddUpdateError(err error) *Updates {
	u.resultErr = err
	return u
}

// AddUpdate adds a database update to be returned by the DBUpdates iterator. If
// AddUpdateError has been set, this method will panic.
func (u *Updates) AddUpdate(update *driver.DBUpdate) *Updates {
	if u.resultErr != nil {
		panic("It is invalid to set more updates after AddUpdateError is defined.")
	}
	u.push(&item{item: update})
	return u
}

// AddDelay adds a delay before the next iteration will complete.
func (u *Updates) AddDelay(delay time.Duration) *Updates {
	u.push(&item{delay: delay})
	return u
}

// Final converts the Updates object to a driver.DBUpdates. This method is
// intended for use within WillExecute() to return results.
func (u *Updates) Final() driver.DBUpdates {
	return &driverDBUpdates{Updates: u}
}
