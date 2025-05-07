package kafka

import (
	"math"
	"time"
)

type RetryHandler struct {
	MaxRetries int
	BaseDelay  time.Duration
}

func NewRetryHandler(max int, base time.Duration) *RetryHandler {
	return &RetryHandler{MaxRetries: max, BaseDelay: base}
}

func (r *RetryHandler) HandleRetry(attempt int) time.Duration {
	if attempt >= r.MaxRetries {
		return -1 // signal to send to DLQ
	}
	// Exponential backoff with jitter
	delay := float64(r.BaseDelay) * math.Pow(2, float64(attempt))
	jitter := time.Duration(float64(r.BaseDelay) * 0.1)
	return time.Duration(delay) + jitter
}
