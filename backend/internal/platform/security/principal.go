package security

import (
	"slices"
	"strings"
)

type Principal struct {
	PublicID    string
	SessionID   string
	Roles       []string
	Permissions []string
}

func NormalizeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	slices.Sort(result)
	return result
}
