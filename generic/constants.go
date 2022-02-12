package generic

import "fmt"

var revision string

func GetRevision() string {
	if revision == "" {
		return "canary"
	}

	if len(revision) > 10 {
		return fmt.Sprintf("%10s", revision)
	}

	return revision
}
