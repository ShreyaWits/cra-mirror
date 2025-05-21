package dtos

type ChannelConfig struct {
	Service  string `json:"service" validate:"required"`
	Primary  string `json:"primary" validate:"required"`
	Fallback string `json:"fallback" validate:"required"`
}
