package utils

import (
	"testing"
)

func TestGenerateRedisCacheKey(t *testing.T) {
	tests := []struct {
		namespace string
		key       string
		want      string
	}{
		{"user", "123", "user:123"},
		{"cache", "token", "cache:token"},
		{"", "empty", ":empty"},
		{"empty", "", "empty:"},
		{"", "", ":"},
		{"ns:with:colon", "key", "ns:with:colon:key"},
	}

	for _, tt := range tests {
		got := GenerateRedisCacheKey(tt.namespace, tt.key)
		if got != tt.want {
			t.Errorf("GenerateRedisCacheKey(%q, %q) = %q; want %q", tt.namespace, tt.key, got, tt.want)
		}
	}
}
