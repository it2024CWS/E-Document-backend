package util

import (
	"fmt"

	"github.com/google/uuid"
)

// ParseUUID parses a string into a UUID
func ParseUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, fmt.Errorf("uuid string is empty")
	}
	return uuid.Parse(s)
}
