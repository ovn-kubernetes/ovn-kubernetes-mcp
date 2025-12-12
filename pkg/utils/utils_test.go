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

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		lines     []string
		want      []string
		wantError bool
	}{
		{
			name:      "match pattern with no lines",
			pattern:   "line 2",
			lines:     []string{},
			want:      []string{},
			wantError: false,
		},
		{
			name:      "match pattern",
			pattern:   "line 2",
			lines:     []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:      []string{"line 2"},
			wantError: false,
		},
		{
			name:      "match pattern with multiple matches",
			pattern:   "line 2",
			lines:     []string{"line 1", "line 2", "line 2", "line 4", "line 5"},
			want:      []string{"line 2", "line 2"},
			wantError: false,
		},
		{
			name:      "match pattern with no matches",
			pattern:   "line 6",
			lines:     []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:      []string{},
			wantError: false,
		},
		{
			name:      "match pattern with invalid regex",
			pattern:   "[invalid(",
			lines:     []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:      []string{},
			wantError: true,
		},
		{
			name:      "match pattern with multiple match patterns separated by |",
			pattern:   "line 2|line 4",
			lines:     []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:      []string{"line 2", "line 4"},
			wantError: false,
		},
		{
			name:      "match pattern with multiple match patterns separated by .",
			pattern:   "line 2.*line 4",
			lines:     []string{"line 1", "line 2", "line 3", "line 4", "line 5", "line 2 line 4"},
			want:      []string{"line 2 line 4"},
			wantError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matchedLines, err := MatchPattern(test.pattern, test.lines)
			if err != nil && !test.wantError {
				t.Fatalf("MatchPattern() unexpected error = %v", err)
			}
			if err == nil && test.wantError {
				t.Fatalf("MatchPattern() expected error but got nil")
			}
			if err == nil {
				if !slices.Equal(matchedLines, test.want) {
					t.Fatalf("MatchPattern() got %v, want %v", matchedLines, test.want)
				}
			}
		})
	}
}

func TestHead(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		n     int
		want  []string
	}{
		{
			name:  "head with no lines",
			lines: []string{},
			n:     10,
			want:  []string{},
		},
		{
			name:  "head 2 lines",
			lines: []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			n:     2,
			want:  []string{"line 1", "line 2"},
		},
		{
			name:  "head 0 lines",
			lines: []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			n:     0,
			want:  []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
		},
		{
			name:  "head -1 lines",
			lines: []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			n:     -1,
			want:  []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
		},
		{
			name:  "head 10 lines",
			lines: []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			n:     10,
			want:  []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Head(test.lines, test.n)
			if !slices.Equal(got, test.want) {
				t.Fatalf("Head() got %v, want %v", got, test.want)
			}
		})
	}
}

func TestTail(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		n     int
		want  []string
	}{
		{
			name:  "tail with no lines",
			lines: []string{},
			n:     10,
			want:  []string{},
		},
		{
			name:  "tail 2 lines",
			lines: []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			n:     2,
			want:  []string{"line 4", "line 5"},
		},
		{
			name:  "tail 0 lines",
			lines: []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			n:     0,
			want:  []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
		},
		{
			name:  "tail -1 lines",
			lines: []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			n:     -1,
			want:  []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
		},
		{
			name:  "tail 10 lines",
			lines: []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			n:     10,
			want:  []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Tail(test.lines, test.n)
			if !slices.Equal(got, test.want) {
				t.Fatalf("Tail() got %v, want %v", got, test.want)
			}
		})
	}
}
