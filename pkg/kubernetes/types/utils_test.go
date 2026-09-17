// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
	"time"
)

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name     string
		age      time.Duration
		expected string
	}{
		{"1s", time.Second, "1s"},
		{"1m0s", time.Minute, "1m0s"},
		{"1h0m", time.Hour, "1h0m"},
		{"1d0h", time.Hour * 24, "1d0h"},
		{"1m4s", 64 * time.Second, "1m4s"},
		{"1h4m", 64 * time.Minute, "1h4m"},
		{"2d16h", 64 * time.Hour, "2d16h"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FormatAge(test.age)
			if result != test.expected {
				t.Errorf("FormatAge(%v) = %v, expected %v", test.age, result, test.expected)
			}
		})
	}
}
