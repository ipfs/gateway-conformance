package test

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"
)

const SKIPS_ENV_VAR = "TEST_SKIPS"

var skips = []string{}

func GetSkips() []string {
	skips := os.Getenv(SKIPS_ENV_VAR)
	if skips == "" {
		return []string{}
	}

	var skipsList []string
	err := json.Unmarshal([]byte(skips), &skipsList)
	if err != nil {
		panic(err)
	}

	return skipsList
}

func split(name string) []string {
	// split name by "/" not prefixed with "\"
	parts := []string{}
	current := ""
	hasSlash := false

	for _, c := range name {
		if hasSlash {
			if c == '/' || c == '\\' {
				current += string(c)
				hasSlash = false
			} else {
				current += "\\"
				current += string(c)
				hasSlash = false
			}
			continue
		}

		if c == '/' {
			parts = append(parts, current)
			current = ""
			continue
		}
		if c == '\\' {
			hasSlash = true
			continue
		}

		current += string(c)
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

func isSkipped(name string, skips []string) bool {
	for _, skip := range skips {
		if name == skip {
			return true
		}

		skipParts := split(skip)
		nameParts := split(name)

		if len(skipParts) > len(nameParts) {
			continue
		}

		matches := true
		for i := range skipParts {
			skipPart := skipParts[i]
			skipPart = "^" + skipPart + "$"

			matched, err := regexp.MatchString(skipPart, nameParts[i])
			if err != nil {
				panic(err)
			}

			if !matched {
				matches = false
				break
			}
		}

		return matches
	}

	return false
}

func MustNotBeSkipped(t *testing.T) {
	t.Helper()
	skipped := isSkipped(t.Name(), skips)

	if skipped {
		t.Skipf("skipped")
	}
}

// init will load the skips
func init() {
	skips = GetSkips()
}
