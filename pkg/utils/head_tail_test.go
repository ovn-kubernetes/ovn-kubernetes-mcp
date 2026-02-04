package utils

import (
	"slices"
	"testing"
)

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
			got := head(test.lines, test.n)
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
			got := tail(test.lines, test.n)
			if !slices.Equal(got, test.want) {
				t.Fatalf("Tail() got %v, want %v", got, test.want)
			}
		})
	}
}

func TestApply(t *testing.T) {
	defaultMaxLines := 2
	tests := []struct {
		name           string
		head           int
		tail           int
		applyTailFirst bool
		lines          []string
		want           []string
	}{
		{
			name:           "apply with no params",
			head:           0,
			tail:           0,
			applyTailFirst: false,
			lines:          []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:           []string{"line 1", "line 2"},
		},
		{
			name:           "apply with head",
			head:           3,
			tail:           0,
			applyTailFirst: false,
			lines:          []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:           []string{"line 1", "line 2", "line 3"},
		},
		{
			name:           "apply with tail",
			head:           0,
			tail:           3,
			applyTailFirst: false,
			lines:          []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:           []string{"line 3", "line 4", "line 5"},
		},
		{
			name:           "apply with both head and tail",
			head:           3,
			tail:           2,
			applyTailFirst: false,
			lines:          []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:           []string{"line 2", "line 3"},
		},
		{
			name:           "apply with both head and tail and apply tail first",
			head:           2,
			tail:           3,
			applyTailFirst: true,
			lines:          []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:           []string{"line 3", "line 4"},
		},
		{
			name:           "apply with both head and tail with tail value greater than head value",
			head:           2,
			tail:           3,
			applyTailFirst: false,
			lines:          []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:           []string{"line 1", "line 2"},
		},
		{
			name:           "apply with both head and tail with head value greater than tail value and apply tail first",
			head:           3,
			tail:           2,
			applyTailFirst: true,
			lines:          []string{"line 1", "line 2", "line 3", "line 4", "line 5"},
			want:           []string{"line 4", "line 5"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params := HeadTailParams{
				Head:           test.head,
				Tail:           test.tail,
				ApplyTailFirst: test.applyTailFirst,
			}
			got := params.Apply(test.lines, defaultMaxLines)
			if !slices.Equal(got, test.want) {
				t.Fatalf("Apply() got %v, want %v", got, test.want)
			}
		})
	}
}
