package utils

// AlwaysRetriable always deems an error retriable.
func AlwaysRetriable(err error) bool {
	return true
}
