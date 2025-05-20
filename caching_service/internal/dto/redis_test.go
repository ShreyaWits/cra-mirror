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
			req:     GetCacheRequest{Namespace: "ns", Key: "key"},
			wantErr: false,
		},
		{
			name:    "missing namespace",
			req:     GetCacheRequest{Namespace: "", Key: "key"},
			wantErr: true,
		},
		{
			name:    "missing key",
			req:     GetCacheRequest{Namespace: "ns", Key: ""},
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
	tests := []struct {
		name    string
		req     SetCacheRequest
		wantErr bool
	}{
		{
			name: "valid",
			req: SetCacheRequest{
				Namespace: "abc",
				Key:       "def",
				Value:     "some value",
				TTL:       10,
			},
			wantErr: false,
		},
		{
			name: "namespace too short",
			req: SetCacheRequest{
				Namespace: "ab",
				Key:       "def",
				Value:     "some value",
				TTL:       10,
			},
			wantErr: true,
		},
		{
			name: "key too short",
			req: SetCacheRequest{
				Namespace: "abc",
				Key:       "de",
				Value:     "some value",
				TTL:       10,
			},
			wantErr: true,
		},
		{
			name: "missing value",
			req: SetCacheRequest{
				Namespace: "abc",
				Key:       "def",
				Value:     "",
				TTL:       10,
			},
			wantErr: true,
		},
		{
			name: "ttl zero",
			req: SetCacheRequest{
				Namespace: "abc",
				Key:       "def",
				Value:     "some value",
				TTL:       0,
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
			req:     DeleteCacheRequest{Namespace: "ns", Key: "key"},
			wantErr: false,
		},
		{
			name:    "missing namespace",
			req:     DeleteCacheRequest{Namespace: "", Key: "key"},
			wantErr: true,
		},
		{
			name:    "missing key",
			req:     DeleteCacheRequest{Namespace: "ns", Key: ""},
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
