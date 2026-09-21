package domain

import (
	"strings"
	"unicode/utf8"
)

// NormalizeConfigName validates a newly supplied name, not historical snapshots.
func NormalizeConfigName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 120 {
		return "", ErrInvalidConfigName
	}
	return name, nil
}
