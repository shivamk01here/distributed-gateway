package limiter

import (
	"testing"
)

func TestTokenBucket_Allow(t *testing.T) {
	tb := NewTokenBucket(3, 1.0)

	testIP := "192.168.1.1"
	otherIP := "10.0.0.1"

	tests := []struct {
		name          string
		ip            string
		expectAllowed bool
	}{
		{"Request 1: Consumes token 1", testIP, true},
		{"Request 2: Consumes token 2", testIP, true},
		{"Request 3: Consumes token 3", testIP, true},
		{"Request 4: Throttled (Bucket is empty)", testIP, false},
		{"Request 5: Throttled (Bucket still empty)", testIP, false},
		{"Different IP: Should be allowed (Fresh Bucket)", otherIP, true},
	}

	for _, tt := range tests {
		// t.Run executes each case as an isolated sub-test
		t.Run(tt.name, func(t *testing.T) {
			allowed := tb.Allow(tt.ip)
			if allowed != tt.expectAllowed {
				t.Errorf("For IP %s, expected Allow() to be %v, but got %v", tt.ip, tt.expectAllowed, allowed)
			}
		})
	}
}
