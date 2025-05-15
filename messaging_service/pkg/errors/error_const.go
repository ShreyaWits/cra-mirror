package errors

const (
	// MSGxxx (Messaging Service)
	MSGErrInvalidRequest = "MSG001"
	MSGErrInvalidTopic   = "MSG002"
	MSGErrInvalidGroup   = "MSG003"
	MSGErrArgument       = "MSG004"

	// PUBxxx (Publish)
	PUBErrInvalidMessage   = "PUB001"
	PUBErrPublishFailed    = "PUB002"
	PUBErrInvalidConfig    = "PUB003"
	PUBErrTopicNotExists   = "PUB004"
	PUBErrProducerNotReady = "PUB005"

	// SUBxxx (Subscribe)
	SUBErrInvalidConfig    = "SUB001"
	SUBErrSubscribeFailed  = "SUB002"
	SUBErrTopicNotExists   = "SUB003"
	SUBErrConsumerNotReady = "SUB004"
	SUBErrInvalidGroupID   = "SUB005"
	SUBErrStreamError      = "SUB006"

	// TOPxxx (Topic Management)
	TOPErrInvalidConfig     = "TOP001"
	TOPErrCreateFailed      = "TOP002"
	TOPErrTopicExists       = "TOP003"
	TOPErrInvalidPartitions = "TOP004"

	// KAFxxx (Kafka Client)
	KAFErrConnectionFailed  = "KAF001"
	KAFErrBrokerUnavailable = "KAF002"
	KAFErrNotAuthorized     = "KAF003"
	KAFErrNotAuthenticated  = "KAF004"
)
