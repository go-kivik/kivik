package memory

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/flimzy/kivik/driver"
)

var errFindNotImplemented = errors.New("find feature not yet implemented")

type query struct {
	Selector map[string]interface{} `json:"selector"`
	Limit    int64                  `json:"limit"`
	Skip     int64                  `json:"skip"`
	Sort     []string               `json:"sort"`
	Fields   []string               `json:"fields"`
	UseIndex indexSpec              `json:"use_index"`
}

type indexSpec struct {
	ddoc  string
	index string
}

func (i *indexSpec) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		return json.Unmarshal(data, &i.ddoc)
	}
	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	if len(values) == 0 || len(values) > 2 {
		return errors.New("invalid index specification")
	}
	i.ddoc = values[0]
	if len(values) == 2 {
		i.index = values[1]
	}
	return nil
}

func (d *db) Find(_ context.Context, query interface{}) (driver.Rows, error) {
	return nil, nil
}

func (d *db) CreateIndex(_ context.Context, ddoc, name string, index interface{}) error {
	return errFindNotImplemented
}

func (d *db) GetIndexes(_ context.Context) ([]driver.Index, error) {
	return nil, errFindNotImplemented
}

func (d *db) DeleteIndex(_ context.Context, ddoc, name string) error {
	return errFindNotImplemented
}
