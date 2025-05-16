package enums

import "strings"

type KeyType string

const (
	KeyPrivate KeyType = "private"
	KeyPublic  KeyType = "public"
)

// ValidKeyTypes holds all known KeyType values
var ValidKeyTypes = []KeyType{KeyPrivate, KeyPublic}

// HasKeyTypeIdentifier checks if the given value starts with a known KeyType identifier
func HasKeyTypeIdentifier(value string) bool {
	for _, keyType := range ValidKeyTypes {
		if strings.HasPrefix(value, string(keyType)+"_") {
			return true
		}
	}
	return false
}

// AddKeyTypeIdentifier adds the KeyType prefix to a value if not already present.
// Returns the result and true if the prefix was added.
func (k KeyType) AddKeyTypeIdentifier(value string) string {
	prefix := string(k) + "_"
	return prefix + value
}

// RemoveKeyTypeIdentifier removes the KeyType prefix from the value.
// Returns the result and true if a known prefix was removed.
func RemoveKeyTypeIdentifier(value string) (string, *KeyType, bool) {
	for _, keyType := range ValidKeyTypes {
		prefix := string(keyType) + "_"
		if strings.HasPrefix(value, prefix) {
			return value[len(prefix):], &keyType, true
		}
	}
	return value, nil, false // No known prefix removed
}

func IsValidKeyType(value string) (bool, KeyType) {
	for _, keyType := range ValidKeyTypes {
		if string(keyType) == value {
			return true, keyType
		}
	}
	return false, ""
}
