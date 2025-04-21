package enums

import "strings"

type KeyType string

const (
	KMSPrivate KeyType = "private"
	KMSPublic  KeyType = "public"
)

// ValidKMSTypes holds all known KMSType values
var ValidKMSTypes = []KeyType{KMSPrivate, KMSPublic}

// HasKMSTypeIdentifier checks if the given value starts with a known KMSType identifier
func HasKMSTypeIdentifier(value string) bool {
	for _, kmsType := range ValidKMSTypes {
		if strings.HasPrefix(value, string(kmsType)+"_") {
			return true
		}
	}
	return false
}

// AddKMSTypeIdentifier adds the KMSType prefix to a value if not already present.
// Returns the result and true if the prefix was added.
func (k KeyType) AddKMSTypeIdentifier(value string) string {
	prefix := string(k) + "_"
	return prefix + value
}

// RemoveKMSTypeIdentifier removes the KMSType prefix from the value.
// Returns the result and true if a known prefix was removed.
func RemoveKMSTypeIdentifier(value string) (string, *KeyType, bool) {
	for _, kmsType := range ValidKMSTypes {
		prefix := string(kmsType) + "_"
		if strings.HasPrefix(value, prefix) {
			return value[len(prefix):], &kmsType, true
		}
	}
	return value, nil, false // No known prefix removed
}

func IsValidKMSType(value string) (bool, KeyType) {
	for _, kmsType := range ValidKMSTypes {
		if string(kmsType) == value {
			return true, kmsType
		}
	}
	return false, ""
}
