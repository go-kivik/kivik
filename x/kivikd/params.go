package kivikd

import (
	"net/http"
)

// StringQueryParam extracts a query parameter as string.
func StringQueryParam(r *http.Request, key string) (string, bool) {
	values := r.URL.Query()
	if _, ok := values[key]; !ok {
		return "", false
	}
	return values.Get(key), true
}
