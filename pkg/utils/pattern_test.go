package utils

import (
	"slices"
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name       string
		checkMatch bool
		pattern    string
		lines      []string
		want       []string
		wantError  bool
	}{
		{
			name:       "match pattern with no lines",
			checkMatch: true,
			pattern:    "line 2",
			lines:      []string{},
			want:       []string{},
			wantError:  false,
		},
		{
			name:       "match pattern",
			checkMatch: true,
			pattern:    "line 2",
			lines:      []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:       []string{"line 2"},
			wantError:  false,
		},
		{
			name:       "match pattern with multiple matches",
			checkMatch: true,
			pattern:    "line 2",
			lines:      []string{"line 1", "line 2", "line 2", "line 4", "line 5"},
			want:       []string{"line 2", "line 2"},
			wantError:  false,
		},
		{
			name:       "match pattern with no matches",
			checkMatch: true,
			pattern:    "line 6",
			lines:      []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:       []string{},
			wantError:  false,
		},
		{
			name:       "match pattern with invalid regex",
			checkMatch: true,
			pattern:    "[invalid(",
			lines:      []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:       []string{},
			wantError:  true,
		},
		{
			name:       "match pattern with multiple match patterns separated by |",
			checkMatch: true,
			pattern:    "line 2|line 4",
			lines:      []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:       []string{"line 2", "line 4"},
			wantError:  false,
		},
		{
			name:       "match pattern with multiple match patterns separated by .",
			checkMatch: true,
			pattern:    "line 2.*line 4",
			lines:      []string{"line 1", "line 2", "line 3", "line 4", "line 5", "line 2 line 4"},
			want:       []string{"line 2 line 4"},
			wantError:  false,
		},
		{
			name:       "match pattern with no check match",
			checkMatch: false,
			pattern:    "line 2",
			lines:      []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:       []string{},
			wantError:  true,
		},
		{
			name:       "no pattern matching and no check match",
			checkMatch: false,
			pattern:    "",
			lines:      []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:       []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			wantError:  false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			patternParams := PatternParams{Pattern: test.pattern}
			matchedLines, err := patternParams.ExecuteWithMatch(func() ([]string, error) {
				return test.lines, nil
			}, test.checkMatch)
			if err != nil && !test.wantError {
				t.Fatalf("PatternParams.Match() unexpected error = %v", err)
			}
			if err == nil && test.wantError {
				t.Fatalf("PatternParams.Match() expected error but got nil")
			}
			if err == nil {
				if !slices.Equal(matchedLines, test.want) {
					t.Fatalf("PatternParams.Match() got %v, want %v", matchedLines, test.want)
				}
			}
		})
	}
}
