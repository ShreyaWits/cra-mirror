package enum

// RequestType enum using string values
type RequestType string

const (
	RequestTypeAuth    RequestType = "auth"
	RequestTypeFile    RequestType = "file"
	RequestTypePayment RequestType = "payment"
)

// GetValidRequestTypes returns the list of valid RequestType values
func GetValidRequestTypes() []string {
	return []string{
		string(RequestTypeAuth),
		string(RequestTypeFile),
		string(RequestTypePayment),
	}
}


type ChannelType string

const (
	ChannelTypeEmail    ChannelType = "email"
	ChannelTypePhone    ChannelType = "phone"
	
)

// GetValidModelTypes returns the list of valid ModelType values
func GetChannelTypes() []string {
	return []string{
		string(ChannelTypeEmail),
		string(ChannelTypePhone),
	
	}
}