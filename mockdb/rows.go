package kivikmock

import (
	"context"
	"time"

	"github.com/go-kivik/kivik/v4/driver"
)

// Rows is a mocked collection of rows.
type Rows struct {
	iter
	offset    int64
	updateSeq string
	totalRows int64
	warning   string
}

func coalesceRows(rows *Rows) *Rows {
	if rows != nil {
		return rows
	}
	return &Rows{}
}

type driverRows struct {
	context.Context
	*Rows
}

var (
	_ driver.Rows       = &driverRows{}
	_ driver.RowsWarner = &driverRows{}
)

func (r *driverRows) Offset() int64     { return r.offset }
func (r *driverRows) UpdateSeq() string { return r.updateSeq }
func (r *driverRows) TotalRows() int64  { return r.totalRows }
func (r *driverRows) Warning() string   { return r.warning }

func (r *driverRows) Next(row *driver.Row) error {
	result, err := r.unshift(r.Context)
	if err != nil {
		return err
	}
	*row = *result.(*driver.Row)
	return nil
}

// CloseError sets an error to be returned when the rows iterator is closed.
func (r *Rows) CloseError(err error) *Rows {
	r.closeErr = err
	return r
}

// Offset sets the offset value to be returned by the rows iterator.
func (r *Rows) Offset(offset int64) *Rows {
	r.offset = offset
	return r
}

// TotalRows sets the total rows value to be returned by the rows iterator.
func (r *Rows) TotalRows(totalRows int64) *Rows {
	r.totalRows = totalRows
	return r
}

// UpdateSeq sets the update sequence value to be returned by the rows iterator.
func (r *Rows) UpdateSeq(seq string) *Rows {
	r.updateSeq = seq
	return r
}

// Warning sets the warning value to be returned by the rows iterator.
func (r *Rows) Warning(warning string) *Rows {
	r.warning = warning
	return r
}

// AddRow adds a row to be returned by the rows iterator. If AddrowError has
// been set, this method will panic.
func (r *Rows) AddRow(row *driver.Row) *Rows {
	if r.resultErr != nil {
		panic("It is invalid to set more rows after AddRowError is defined.")
	}
	r.push(&item{item: row})
	return r
}

// AddRowError adds an error to be returned during row iteration.
func (r *Rows) AddRowError(err error) *Rows {
	r.resultErr = err
	return r
}

// AddDelay adds a delay before the next iteration will complete.
func (r *Rows) AddDelay(delay time.Duration) *Rows {
	r.push(&item{delay: delay})
	return r
}

// Final converts the Rows object to a driver.Rows. This method is intended for
// use within WillExecute() to return results.
func (r *Rows) Final() driver.Rows {
	return &driverRows{Rows: r}
}
