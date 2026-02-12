package weather

import "strings"

// ValidateLocation validates a location string for weather queries
func ValidateLocation(location string) error {
	// Trim whitespace
	location = strings.TrimSpace(location)

	// Check if empty
	if location == "" {
		return ErrInvalidLocation
	}

	// Additional validation rules can be added here
	// For example: minimum length, forbidden characters, etc.

	return nil
}
