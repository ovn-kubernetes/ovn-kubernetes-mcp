package timeout

import (
	"context"
	"testing"
	"time"
)

func TestTimeoutParams_ToDuration(t *testing.T) {
	tests := []struct {
		name    string
		timeout TimeoutParams
		want    time.Duration
	}{
		{name: "timeout is 0", timeout: TimeoutParams{TimeoutSeconds: 0}, want: 0},
		{name: "timeout is 1", timeout: TimeoutParams{TimeoutSeconds: 1}, want: 1 * time.Second},
		{name: "timeout is 10", timeout: TimeoutParams{TimeoutSeconds: 10}, want: 10 * time.Second},
		{name: "timeout is max", timeout: TimeoutParams{TimeoutSeconds: uint32(MaxTimeout.Seconds())}, want: MaxTimeout},
		{name: "timeout is just over max", timeout: TimeoutParams{TimeoutSeconds: uint32(MaxTimeout.Seconds()) + 1}, want: MaxTimeout},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.timeout.ToDuration()
			if got != test.want {
				t.Fatalf("expected %v, but got %v", test.want, got)
			}
		})
	}
}

func TestTimeoutParams_WithTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout TimeoutParams
	}{
		{name: "timeout is 0", timeout: TimeoutParams{TimeoutSeconds: 0}},
		{name: "timeout is 1", timeout: TimeoutParams{TimeoutSeconds: 1}},
		{name: "timeout is 10", timeout: TimeoutParams{TimeoutSeconds: 10}},
		{name: "timeout is max", timeout: TimeoutParams{TimeoutSeconds: uint32(MaxTimeout.Seconds())}},
		{name: "timeout is just over max", timeout: TimeoutParams{TimeoutSeconds: uint32(MaxTimeout.Seconds()) + 1}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancel := test.timeout.WithTimeout(ctx)
			if cancel != nil {
				defer cancel()
			}
			if cancel != nil {
				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("expected context to have a deadline")
				}

				deadline, ok := ctx.Deadline()
				if !ok {
					t.Fatal("expected context to have a deadline, but it doesn't")
				}

				if time.Until(deadline) > test.timeout.ToDuration() {
					t.Fatal("expected context to have a deadline within the timeout duration")
				}
			} else {
				if _, ok := ctx.Deadline(); ok {
					t.Fatal("expected context to not have a deadline, but it does")
				}
			}
		})
	}
}
