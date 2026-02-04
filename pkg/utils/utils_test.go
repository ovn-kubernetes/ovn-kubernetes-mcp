package utils

import (
	"slices"
	"testing"
)

func TestStripEmptyLines(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  []string
	}{
		{
			name:  "strip empty lines with no lines",
			lines: []string{},
			want:  []string{},
		},
		{
			name:  "strip empty lines with no empty lines",
			lines: []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:  []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
		},
		{
			name:  "strip empty lines with empty lines",
			lines: []string{"", "", "", "", "", "", "", "", "", ""},
			want:  []string{},
		},
		{
			name:  "strip empty lines with empty lines and non empty lines",
			lines: []string{"line 1", "line 2", "line 3", "line 4", "line 5", "", "", "line 8", "line 9", "line 10"},
			want:  []string{"line 1", "line 2", "line 3", "line 4", "line 5", "line 8", "line 9", "line 10"},
		},
		{
			name:  "strip whitespace-only lines",
			lines: []string{"line 1", "   ", "\t", "line 2"},
			want:  []string{"line 1", "line 2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := StripEmptyLines(test.lines)
			if !slices.Equal(got, test.want) {
				t.Fatalf("StripEmptyLines() got %v, want %v", got, test.want)
			}
		})
	}
}
