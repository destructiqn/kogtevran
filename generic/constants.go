package generic

import (
	"fmt"
	"os"
)

var (
	revision    string
	environment = os.Getenv("KV_ENVIRONMENT")
)

func GetRevision() string {
	if revision == "" {
		return "canary"
	}

	if len(revision) > 10 {
		return fmt.Sprintf("%10s", revision)
	}

	return revision
}

func GetEnvironment() string {
	if environment == "" {
		return "development"
	}

	return environment
}

func IsDevelopmentEnvironment() bool {
	return GetEnvironment() == "development"
}
