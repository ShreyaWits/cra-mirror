package dto

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

var validate = validator.New()

func TestGetCacheRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     GetCacheRequest
		wantErr bool
	}{
		{
			name:    "valid",
			req:     GetCacheRequest{Namespace: "ns", Key: "key", TrackingId: "123e4567-e89b-42d3-a456-426614174000"},
			wantErr: false,
		},
		{
			name:    "missing namespace",
			req:     GetCacheRequest{Namespace: "", Key: "key", TrackingId: "123e4567-e89b-42d3-a456-426614174000"},
			wantErr: true,
		},
		{
			name:    "missing key",
			req:     GetCacheRequest{Namespace: "ns", Key: "", TrackingId: "123e4567-e89b-42d3-a456-426614174000"},
			wantErr: true,
		},
		{
			name:    "missing tracking id",
			req:     GetCacheRequest{Namespace: "ns", Key: "key", TrackingId: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		err := validate.Struct(tt.req)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestSetCacheRequest_Validation(t *testing.T) {
	validUUIDv4 := "123e4567-e89b-42d3-a456-426614174000"
	tests := []struct {
		name    string
		req     SetCacheRequest
		wantErr bool
	}{
		{
			name: "valid",
			req: SetCacheRequest{
				Namespace:  "abc",
				Key:        "def",
				Value:      "some value",
				TTL:        10,
				TrackingId: validUUIDv4,
			},
			wantErr: false,
		},
		{
			name: "namespace too short",
			req: SetCacheRequest{
				Namespace:  "ab",
				Key:        "def",
				Value:      "some value",
				TTL:        10,
				TrackingId: validUUIDv4,
			},
			wantErr: true,
		},
		{
			name: "key too short",
			req: SetCacheRequest{
				Namespace:  "abc",
				Key:        "de",
				Value:      "some value",
				TTL:        10,
				TrackingId: validUUIDv4,
			},
			wantErr: true,
		},
		{
			name: "missing value",
			req: SetCacheRequest{
				Namespace:  "abc",
				Key:        "def",
				Value:      "",
				TTL:        10,
				TrackingId: validUUIDv4,
			},
			wantErr: true,
		},
		{
			name: "ttl zero",
			req: SetCacheRequest{
				Namespace:  "abc",
				Key:        "def",
				Value:      "some value",
				TTL:        0,
				TrackingId: validUUIDv4,
			},
			wantErr: true,
		},
		{
			name: "invalid tracking id",
			req: SetCacheRequest{
				Namespace:  "abc",
				Key:        "def",
				Value:      "some value",
				TTL:        10,
				TrackingId: "not-a-uuid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		err := validate.Struct(tt.req)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestDeleteCacheRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     DeleteCacheRequest
		wantErr bool
	}{
		{
			name:    "valid",
			req:     DeleteCacheRequest{Namespace: "ns", Key: "key", TrackingId: "123e4567-e89b-42d3-a456-426614174000"},
			wantErr: false,
		},
		{
			name:    "missing namespace",
			req:     DeleteCacheRequest{Namespace: "", Key: "key", TrackingId: "123e4567-e89b-42d3-a456-426614174000"},
			wantErr: true,
		},
		{
			name:    "missing key",
			req:     DeleteCacheRequest{Namespace: "ns", Key: "", TrackingId: "123e4567-e89b-42d3-a456-426614174000"},
			wantErr: true,
		},
		{
			name:    "missing tracking id",
			req:     DeleteCacheRequest{Namespace: "ns", Key: "key", TrackingId: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		err := validate.Struct(tt.req)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}
