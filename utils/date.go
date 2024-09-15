package utils

import (
	_ "golang.org/x/text/message"
	"time"
)

func ConvertUnixPtrToTimePtr(value *int) *time.Time {
	if value == nil {
		return nil
	}
	return ToP(
		time.Unix(int64(*value), 0),
	)
}
