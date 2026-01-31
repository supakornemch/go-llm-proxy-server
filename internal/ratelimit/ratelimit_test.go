package ratelimit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/supakornemchananon/go-llm-proxy-server/internal/ratelimit"
)

func TestLimiter_AllowTPS(t *testing.T) {
	manager := ratelimit.NewManager()

	// Create a limiter allowing 2 TPS
	limiter := manager.GetLimiter("test-key", 2, 0)

	// First 2 requests should pass immediately
	assert.True(t, limiter.AllowTPS())
	assert.True(t, limiter.AllowTPS())
	assert.True(t, limiter.AllowTPS())
}

func TestLimiter_TokenLimit(t *testing.T) {
	manager := ratelimit.NewManager()

	// 6000 tokens per minute = 100 tokens per second.
	// We init with tokenLimit = 6000.
	limiter := manager.GetLimiter("token-key", 0, 6000)

	// Consume 100 tokens
	assert.True(t, limiter.AllowTokens(100))

	// Consume more than available (bucket size is tokenLimit = 6000)
	// If we ask for 6001, it should fail immediately (assuming full bucket)
	assert.False(t, limiter.AllowTokens(6001))
}
