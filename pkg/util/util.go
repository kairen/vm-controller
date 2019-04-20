package util

import (
	"time"
)

func SubtractTime(t time.Time) time.Duration {
	now := time.Now()
	then := now.Sub(t)
	return then
}
