package common

import (
	"os"
)

// DebugOn returns true if the DEBUG environment variable is set to true or 1.
func DebugOn() bool {
	DEBUG := os.Getenv("DEBUG")
	return DEBUG == "true" || DEBUG == "1"
}

func ProdMode() bool {
	PROD := os.Getenv("PROD")
	return PROD == "true" || PROD == "1"
}

// DevMode returns true if not in production.
func DevMode() bool {
	return !ProdMode()
}
