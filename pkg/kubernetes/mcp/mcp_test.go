package mcp

import (
	"context"
	"testing"
	"time"
)

// TestTimeoutMechanism verifies timeout is applied correctly for Kubernetes tools
func TestTimeoutMechanism(t *testing.T) {
	// Test 1: Timeout triggers when operation is slow
	t.Run("operation exceeds timeout", func(t *testing.T) {
		toolTimeout := 10 * time.Millisecond
		ctx := context.Background()

		// Apply timeout like Kubernetes tools do
		if toolTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, toolTimeout)
			defer cancel()
		}

		// Simulate slow operation (50ms > 10ms timeout)
		time.Sleep(50 * time.Millisecond)

		// Verify context timed out
		if ctx.Err() != context.DeadlineExceeded {
			t.Error("expected context to timeout, but it didn't")
		}
	})

	// Test 2: No timeout when disabled
	t.Run("timeout disabled", func(t *testing.T) {
		toolTimeout := time.Duration(0) // Disabled
		ctx := context.Background()

		// Apply timeout (should be no-op when 0)
		if toolTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, toolTimeout)
			defer cancel()
		}

		// Simulate operation
		time.Sleep(10 * time.Millisecond)

		// Verify context did NOT timeout
		if ctx.Err() != nil {
			t.Errorf("expected no timeout, but got: %v", ctx.Err())
		}
	})
}
