package serve

import (
	"net/http"
	"strconv"

	"github.com/flimzy/kivik/errors"
)

func intParam(r *http.Request, key string) (int, bool, error) {
	value, ok := stringParam(r, key)
	if !ok {
		return 0, false, nil
	}
	iValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, true, errors.Statusf(http.StatusBadRequest, "%s parameter must be an integer", key)
	}
	return iValue, true, nil
}

func stringParam(r *http.Request, key string) (string, bool) {
	values := r.URL.Query()
	if _, ok := values[key]; !ok {
		return "", false
	}
	return values.Get(key), true
}
