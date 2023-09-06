package kivikmock

import (
	"fmt"
	"time"
)

func optionsString(opt map[string]interface{}) string {
	if opt == nil {
		return "\n\t- has any options"
	}
	return fmt.Sprintf("\n\t- has options: %v", opt)
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("\n\t- should return error: %s", err)
}

func delayString(delay time.Duration) string {
	if delay == 0 {
		return ""
	}
	return fmt.Sprintf("\n\t- should delay for: %s", delay)
}

func fieldString(field, value string) string {
	if value == "" {
		return "\n\t- has any " + field
	}
	return "\n\t- has " + field + ": " + value
}
