package kafka

type ConsumerMode string

type KafkaConfig struct {
	Brokers      []string
	Topic        string
	GroupID      string
	Mode         ConsumerMode
	BalancerType string
	MinBytes     int
	MaxBytes     int
}
