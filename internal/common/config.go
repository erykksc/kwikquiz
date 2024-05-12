package common

import (
	"os"
)

// DebugOn returns true if the DEBUG environment variable is set to true or 1.
func DebugOn() bool {
	DEBUG := os.Getenv("DEBUG")
	return DEBUG == "true" || DEBUG == "1"
}
