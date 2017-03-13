package serve

import (
	"net/http"
	"strconv"

	"github.com/flimzy/kivik/errors"
)

func intQueryParam(r *http.Request, key string) (int64, bool, error) {
	value, ok := StringQueryParam(r, key)
	if !ok {
		return 0, false, nil
	}
	iValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, true, errors.Statusf(http.StatusBadRequest, "%s parameter must be an integer", key)
	}
	return iValue, true, nil
}

// StringQueryParam extracts a query paramter as string.
func StringQueryParam(r *http.Request, key string) (string, bool) {
	values := r.URL.Query()
	if _, ok := values[key]; !ok {
		return "", false
	}
	return values.Get(key), true
}

func stringParam(r *http.Request, key string) (string, bool) {
	params := getParams(r)
	value, ok := params[key]
	return value, ok
}
