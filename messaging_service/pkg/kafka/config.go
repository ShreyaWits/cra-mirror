package kafka

type ConsumerMode string

const (
	CompetingConsumer ConsumerMode = "competing"
	FanOut            ConsumerMode = "fanout"
)

type KafkaConfig struct {
	Brokers      []string
	Topic        string
	GroupID      string
	Mode         ConsumerMode
	BalancerType string
	MinBytes     int
	MaxBytes     int
}

