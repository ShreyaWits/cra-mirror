// kafka/balancer.go
package kafka

import (
	"strings"

	kafka "github.com/segmentio/kafka-go"
)

func GetBalancer(balancerType string) kafka.Balancer {
	switch strings.ToLower(balancerType) {
	case "roundrobin":
		return &kafka.RoundRobin{}
	case "crc32":
		return &kafka.CRC32Balancer{}
	case "hash":
		return &kafka.Hash{}
	default:
		return &kafka.LeastBytes{}
	}
}
