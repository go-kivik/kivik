package kivikmock

import (
	"github.com/go-kivik/kivik/v4/driver"
)

type driverDB struct {
	*DB
}

var (
	_ driver.DB         = &driverDB{}
	_ driver.BulkGetter = &driverDB{}
	_ driver.Finder     = &driverDB{}
)

func (db *driverDB) Close() error {
	expected := &ExpectedDBClose{
		commonExpectation: commonExpectation{db: db.DB},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback()
	}
	return expected.err
}
