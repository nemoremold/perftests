package utils

import "time"

// GetDurationSince receives a time and calculates the timelapse since that time till now.
func GetDurationSince(previous time.Time) time.Duration {
	return time.Since(previous)
}
