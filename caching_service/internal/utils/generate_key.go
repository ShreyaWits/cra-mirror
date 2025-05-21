package utils

import "fmt"

func GenerateRedisCacheKey(Namespace, Key string) string {
	return fmt.Sprintf("%s:%s", Namespace, Key)
}
