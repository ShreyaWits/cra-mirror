package dto

type GetCacheRequest struct {
	Namespace  string `json:"namespace" validate:"required"`
	Key        string `json:"key" validate:"required"`
	TrackingId string `json:"tracking_id" validate:"required,uuid4"`
}

type SetCacheRequest struct {
	Namespace  string `json:"namespace" validate:"required,min=3"`
	Key        string `json:"key" validate:"required,min=3"`
	Value      string `json:"value" validate:"required"`
	TTL        int64  `json:"ttl" validate:"required,gt=0"`
	TrackingId string `json:"tracking_id" validate:"required,uuid4"`
}

type DeleteCacheRequest struct {
	Namespace  string `json:"namespace" validate:"required"`
	Key        string `json:"key" validate:"required"`
	TrackingId string `json:"tracking_id" validate:"required,uuid4"`
}
