package services

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "nps-reciept-service/proto"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type mockRedisClient struct {
	incr   func(ctx context.Context, key string) *redis.IntCmd
	expire func(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}

func (m *mockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	return m.incr(ctx, key)
}

func (m *mockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return m.expire(ctx, key, expiration)
}

func TestClaimServer_ProcessClaim(t *testing.T) {
	t.Run("Valid Claim Request", func(t *testing.T) {
		mock := &mockRedisClient{
			incr: func(ctx context.Context, key string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetVal(1)
				return cmd
			},
			expire: func(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
				cmd := redis.NewBoolCmd(ctx)
				cmd.SetVal(true)
				return cmd
			},
		}

		server := NewClaimServer(mock)
		req := &pb.ClaimRequest{
			Pran:        "123456789012",
			DateOfClaim: "2024-04-24",
		}

		resp, err := server.ProcessClaim(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		assert.Contains(t, resp.Data.ClaimId, "CLM2404249012")
	})

	t.Run("Non-First Claim Request (sequence > 1)", func(t *testing.T) {
		mock := &mockRedisClient{
			incr: func(ctx context.Context, key string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetVal(5) // Not the first sequence
				return cmd
			},
			expire: func(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
				// This should not be called when sequence > 1
				t.Errorf("Expire should not be called when sequence > 1")
				cmd := redis.NewBoolCmd(ctx)
				cmd.SetVal(true)
				return cmd
			},
		}

		server := NewClaimServer(mock)
		req := &pb.ClaimRequest{
			Pran:        "123456789012",
			DateOfClaim: "2024-04-24",
		}

		resp, err := server.ProcessClaim(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		assert.Contains(t, resp.Data.ClaimId, "CLM2404249012")
		assert.Equal(t, "CLM24042490120005", resp.Data.ClaimId) // Sequence should be 0005
	})

	t.Run("Invalid PRAN", func(t *testing.T) {
		server := NewClaimServer(nil)
		req := &pb.ClaimRequest{
			Pran:        "123",
			DateOfClaim: "2024-04-24",
		}

		resp, err := server.ProcessClaim(context.Background(), req)
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "CLM0001")
	})

	t.Run("Invalid Date", func(t *testing.T) {
		server := NewClaimServer(nil)
		req := &pb.ClaimRequest{
			Pran:        "123456789012",
			DateOfClaim: "2024/04/24",
		}

		resp, err := server.ProcessClaim(context.Background(), req)
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "CLM0004")
	})

	t.Run("Redis Incr Error", func(t *testing.T) {
		mock := &mockRedisClient{
			incr: func(ctx context.Context, key string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetErr(errors.New("redis error"))
				return cmd
			},
			expire: func(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
				cmd := redis.NewBoolCmd(ctx)
				cmd.SetVal(true)
				return cmd
			},
		}
	
		server := NewClaimServer(mock)
		req := &pb.ClaimRequest{
			Pran:        "123456789012",
			DateOfClaim: "2024-04-24",
		}
	
		resp, err := server.ProcessClaim(context.Background(), req)
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "CLM0005")
	})

	t.Run("Redis Expire Error", func(t *testing.T) {
		mock := &mockRedisClient{
			incr: func(ctx context.Context, key string) *redis.IntCmd {
				cmd := redis.NewIntCmd(ctx)
				cmd.SetVal(1) // First time, so Expire will be called
				return cmd
			},
			expire: func(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
				cmd := redis.NewBoolCmd(ctx)
				cmd.SetErr(errors.New("expire error"))
				return cmd
			},
		}
	
		server := NewClaimServer(mock)
		req := &pb.ClaimRequest{
			Pran:        "123456789012",
			DateOfClaim: "2024-04-24",
		}
	
		// Should still succeed even with expire error as it's handled gracefully
		resp, err := server.ProcessClaim(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		assert.Contains(t, resp.Data.ClaimId, "CLM2404249012")
	})
}

func TestFormatSequence(t *testing.T) {
	tests := []struct {
		sequence int64
		expected string
	}{
		{1, "0001"},
		{42, "0042"},
		{123, "0123"},
		{9999, "9999"},
		{10000, "10000"}, // Edge case: overflow
	}

	for _, tc := range tests {
		t.Run("Sequence_"+tc.expected, func(t *testing.T) {
			result := formatSequence(tc.sequence)
			assert.Equal(t, tc.expected, result)
		})
	}
}