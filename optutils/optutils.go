package optutils

import (
	"fmt"
	"strconv"
)

type optErr string

var _ error = optErr("")

func (e optErr) Error() string {
	return string(e)
}

// Standard errors
const (
	ErrKeyNotFound      = optErr("key not found")
	ErrOverflow         = optErr("type assertion results in overflow")
	ErrConversionFailed = optErr("type conversion failed")
	ErrInvalidType      = optErr("invalid type for key")
)

// Int64 returns the requested option key as an Int64.
func Int64(opts map[string]interface{}, key string) (int64, error) {
	v, ok := opts[key]
	if !ok {
		return 0, ErrKeyNotFound
	}
	switch t := v.(type) {
	case int:
		return int64(t), nil
	case int8:
		return int64(t), nil
	case int16:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case uint:
		return int64(t), nil
	case uint8:
		return int64(t), nil
	case uint16:
		return int64(t), nil
	case uint32:
		return int64(t), nil
	case uint64:
		newV := int64(t)
		fmt.Printf("new = %d old = %d", newV, t)
		if uint64(newV) != t { // overflow check
			return 0, ErrOverflow
		}
		return newV, nil
	case int64:
		return t, nil
	case string:
		newV, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return 0, ErrConversionFailed
		}
		return newV, nil
	}
	return 0, ErrInvalidType
}
