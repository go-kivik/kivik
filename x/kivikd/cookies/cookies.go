package cookies

import (
	"bytes"
	"encoding/base64"
	"strconv"

	"github.com/pkg/errors"
)

// DecodeCookie decodes a Base64-encoded cookie, and returns its component
// parts.
func DecodeCookie(cookie string) (name string, created int64, err error) {
	data, err := base64.RawURLEncoding.DecodeString(cookie)
	if err != nil {
		return "", 0, err
	}
	parts := bytes.SplitN(data, []byte(":"), 3)
	t, err := strconv.ParseInt(string(parts[1]), 16, 64)
	if err != nil {
		return "", 0, errors.Wrap(err, "invalid timestamp")
	}
	return string(parts[0]), t, nil
}
