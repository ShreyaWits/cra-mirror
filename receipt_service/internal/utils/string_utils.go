package utils

import "unicode"

// IsNumeric checks if a string contains only numeric digits.
func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}