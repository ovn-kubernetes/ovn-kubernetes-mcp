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
