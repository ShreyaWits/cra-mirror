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

	// CFGxxx (Configuration Management)
	CFGErrFetchFailed      = "CFG001"
	CFGErrValidationFailed = "CFG002"
	CFGErrMarshalFailed    = "CFG003"
	CFGErrUnmarshalFailed  = "CFG004"

	// CACxxx (Cache Operations)
	CACErrConnectionFailed = "CAC001"
	CACErrSetFailed        = "CAC002"
	CACErrGetFailed        = "CAC003"
	CACErrNotFound         = "CAC004"
	CACErrInvalidateFailed = "CAC005"
)
