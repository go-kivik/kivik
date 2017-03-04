package serve

import (
	"net/http"
	"strconv"

	"github.com/flimzy/kivik/errors"
)

func intParam(r *http.Request, key string) (int, bool, error) {
	params := getParams(r)
	value, ok := params[key]
	if !ok {
		return 0, false, nil
	}
	iValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, true, errors.Status(http.StatusBadRequest, "bytes parameter must be an integer")
	}
	return iValue, true, nil
}
