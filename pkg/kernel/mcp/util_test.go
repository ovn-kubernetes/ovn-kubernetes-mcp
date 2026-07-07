package mcp

import (
	"testing"
)

func TestFilterWarnings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single line without warning",
			input:    "This is a normal line",
			expected: "This is a normal line",
		},
		{
			name:     "single line with Warning at start",
			input:    "Warning: something happened",
			expected: "",
		},
		{
			name:     "single line with Warning in middle",
			input:    "Error: Warning detected in system",
			expected: "",
		},
		{
			name:     "single line with lowercase warning",
			input:    "This is a warning message",
			expected: "",
		},
		{
			name:     "multiple lines without warnings",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "multiple lines with one warning",
			input:    "Line 1\nWarning: something wrong\nLine 3",
			expected: "Line 1\nLine 3",
		},
		{
			name:     "multiple lines with multiple warnings",
			input:    "Line 1\nWarning: first warning\nLine 3\nWarning: second warning\nLine 5",
			expected: "Line 1\nLine 3\nLine 5",
		},
		{
			name:     "line with Warning substring",
			input:    "ReWarning test\nWarningLevel\nLine without",
			expected: "Line without",
		},
		{
			name:     "empty lines mixed with warnings",
			input:    "\nWarning: test\n\nNormal line\n",
			expected: "\n\nNormal line\n",
		},
		{
			name:     "warning with special characters",
			input:    "Warning: file.txt not found!\nProcessed successfully",
			expected: "Processed successfully",
		},
		{
			name:     "case sensitive - warning vs Warning vs WARNING",
			input:    "warning: lowercase\nWarning: uppercase\nWARNING: all caps",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterWarnings(tt.input)
			if result != tt.expected {
				t.Errorf("filterWarnings() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSplitConntrackSummary(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedSummary   string
		expectedRemaining string
	}{
		{
			name:              "empty string",
			input:             "",
			expectedSummary:   "",
			expectedRemaining: "",
		},
		{
			name:              "conntrack dump summary",
			input:             "conntrack v1.4.8 (conntrack-tools): 461 flow entries have been shown.",
			expectedSummary:   "conntrack v1.4.8 (conntrack-tools): 461 flow entries have been shown.",
			expectedRemaining: "",
		},
		{
			name:              "conntrack dump summary with other stderr",
			input:             "conntrack v1.4.9 (conntrack-tools): 52 flow entries have been shown.\nreal error message",
			expectedSummary:   "conntrack v1.4.9 (conntrack-tools): 52 flow entries have been shown.",
			expectedRemaining: "real error message",
		},
		{
			name:              "unrelated stderr preserved",
			input:             "command not found",
			expectedSummary:   "",
			expectedRemaining: "command not found",
		},
		{
			name:              "partial match not treated as summary",
			input:             "461 flow entries have been shown.",
			expectedSummary:   "",
			expectedRemaining: "461 flow entries have been shown.",
		},
		{
			name:              "non conntrack-tools package not treated as summary",
			input:             "conntrack v1.4.8 (other-tools): 10 flow entries have been shown.",
			expectedSummary:   "",
			expectedRemaining: "conntrack v1.4.8 (other-tools): 10 flow entries have been shown.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, remaining := splitConntrackSummary(tt.input)
			if summary != tt.expectedSummary {
				t.Errorf("splitConntrackSummary() summary = %q, want %q", summary, tt.expectedSummary)
			}
			if remaining != tt.expectedRemaining {
				t.Errorf("splitConntrackSummary() remaining = %q, want %q", remaining, tt.expectedRemaining)
			}
		})
	}
}
