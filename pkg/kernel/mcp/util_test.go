package kernel

import (
	"fmt"
	"slices"
	"strings"
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

func TestValidateParameters(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		wantError bool
	}{
		{
			name:      "empty string",
			param:     "",
			wantError: false,
		},
		{
			name:      "valid with hyphens",
			param:     "--param",
			wantError: false,
		},
		{
			name:      "valid with equals",
			param:     "key=value",
			wantError: false,
		},
		{
			name:      "valid with spaces",
			param:     "hello world",
			wantError: false,
		},
		{
			name:      "invalid with semicolon",
			param:     "test;rm -rf",
			wantError: true,
		},
		{
			name:      "invalid with ampersand",
			param:     "test&background",
			wantError: true,
		},
		{
			name:      "invalid with pipe",
			param:     "test|grep",
			wantError: true,
		},
		{
			name:      "invalid with dollar sign",
			param:     "test$var",
			wantError: true,
		},
		{
			name:      "invalid with backtick",
			param:     "test`whoami`",
			wantError: true,
		},
		{
			name:      "invalid with less than",
			param:     "test<input",
			wantError: true,
		},
		{
			name:      "invalid with greater than",
			param:     "test>output",
			wantError: true,
		},
		{
			name:      "invalid with backslash",
			param:     "test\\escape",
			wantError: true,
		},
		{
			name:      "invalid with opening parenthesis",
			param:     "test(subshell",
			wantError: true,
		},
		{
			name:      "invalid with closing parenthesis",
			param:     "test)subshell",
			wantError: true,
		},
		{
			name:      "invalid with multiple metacharacters",
			param:     "test;ls|grep&",
			wantError: true,
		},
		{
			name:      "invalid with semicolon only",
			param:     ";",
			wantError: true,
		},
		{
			name:      "invalid with command substitution",
			param:     "$(whoami)",
			wantError: true,
		},
		{
			name:      "invalid with backtick command substitution",
			param:     "`id`",
			wantError: true,
		},
		{
			name:      "valid IP address",
			param:     "192.168.1.1",
			wantError: false,
		},
		{
			name:      "valid CIDR notation",
			param:     "10.0.0.0/24",
			wantError: false,
		},
		{
			name:      "valid with port number",
			param:     "127.0.0.1:8080",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParameters(tt.param)
			if tt.wantError && err == nil {
				t.Errorf("validateParameters(%q) expected error, got nil", tt.param)
			}
			if !tt.wantError && err != nil {
				t.Errorf("validateParameters(%q) expected no error, got %v", tt.param, err)
			}
		})
	}
}

func TestLimitOutputLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLines int
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			maxLines: 10,
			expected: "",
		},
		{
			name:     "single line within limit",
			input:    "This is a single line",
			maxLines: 10,
			expected: "This is a single line",
		},
		{
			name:     "multiple lines within limit",
			input:    "Line 1\nLine 2\nLine 3",
			maxLines: 5,
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "exactly at limit",
			input:    "Line 1\nLine 2\nLine 3",
			maxLines: 3,
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "exceeds limit - simple case",
			input:    "Line 1\nLine 2\nLine 3\nLine 4\nLine 5",
			maxLines: 3,
			expected: "Line 1\nLine 2\nLine 3\n\n... Output truncated. Showing first 3 lines out of 5 total lines.",
		},
		{
			name:     "exceeds limit by one",
			input:    "Line 1\nLine 2\nLine 3\nLine 4",
			maxLines: 3,
			expected: "Line 1\nLine 2\nLine 3\n\n... Output truncated. Showing first 3 lines out of 4 total lines.",
		},
		{
			name:     "exceeds limit by many",
			input:    "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10",
			maxLines: 2,
			expected: "Line 1\nLine 2\n\n... Output truncated. Showing first 2 lines out of 10 total lines.",
		},
		{
			name:     "maxLines is zero - uses default MaxOutputLines",
			input:    "Line 1\nLine 2",
			maxLines: 0,
			expected: "Line 1\nLine 2",
		},
		{
			name:     "maxLines is negative - uses default MaxOutputLines",
			input:    "Line 1\nLine 2",
			maxLines: -5,
			expected: "Line 1\nLine 2",
		},
		{
			name:     "maxLines is 1",
			input:    "Line 1\nLine 2\nLine 3",
			maxLines: 1,
			expected: "Line 1\n\n... Output truncated. Showing first 1 lines out of 3 total lines.",
		},
		{
			name:     "empty lines included in count",
			input:    "Line 1\n\nLine 3\n\nLine 5",
			maxLines: 3,
			expected: "Line 1\n\nLine 3\n\n... Output truncated. Showing first 3 lines out of 5 total lines.",
		},
		{
			name:     "lines with special characters",
			input:    "Line 1 with $special\nLine 2 with &chars\nLine 3 with |pipes\nLine 4",
			maxLines: 2,
			expected: "Line 1 with $special\nLine 2 with &chars\n\n... Output truncated. Showing first 2 lines out of 4 total lines.",
		},
		{
			name:     "lines with tabs and spaces",
			input:    "\tTabbed line\n    Spaced line\nNormal line\nAnother line",
			maxLines: 2,
			expected: "\tTabbed line\n    Spaced line\n\n... Output truncated. Showing first 2 lines out of 4 total lines.",
		},
		{
			name:     "very long single line",
			input:    "This is a very long line that contains lots of text and should still be treated as a single line",
			maxLines: 5,
			expected: "This is a very long line that contains lots of text and should still be treated as a single line",
		},
		{
			name:     "trailing newline preserved when within limit",
			input:    "Line 1\nLine 2\n",
			maxLines: 3,
			expected: "Line 1\nLine 2\n",
		},
		{
			name:     "trailing newline counts as extra line",
			input:    "Line 1\nLine 2\n",
			maxLines: 2,
			expected: "Line 1\nLine 2\n\n... Output truncated. Showing first 2 lines out of 3 total lines.",
		},
		{
			name:     "mixed content types",
			input:    "Header\nData: value1\nData: value2\nData: value3\nData: value4\nFooter",
			maxLines: 4,
			expected: "Header\nData: value1\nData: value2\nData: value3\n\n... Output truncated. Showing first 4 lines out of 6 total lines.",
		},
		{
			name:     "unicode characters",
			input:    "行1\n行2\n行3\n行4\n行5",
			maxLines: 3,
			expected: "行1\n行2\n行3\n\n... Output truncated. Showing first 3 lines out of 5 total lines.",
		},
		{
			name:     "large output with default limit",
			input:    generateLargeOutput(150),
			maxLines: 0, // Should use MaxOutputLines (100)
			expected: generateLargeOutput(100) + "\n\n... Output truncated. Showing first 100 lines out of 150 total lines.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := limitOutputLines(tt.input, tt.maxLines)
			if result != tt.expected {
				t.Errorf("limitOutputLines() =\n%q\nwant\n%q", result, tt.expected)
			}
		})
	}
}

// generateLargeOutput generates n lines of output for testing
func generateLargeOutput(n int) string {
	var lines []string
	for i := 0; i < n; i++ {
		lines = append(lines, fmt.Sprintf("Line %d", i+1))
	}
	return strings.Join(lines, "\n")
}

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name     string
		baseCmd  []string
		expected []string
	}{
		{
			name:     "single base command",
			baseCmd:  []string{"ls"},
			expected: []string{"ls"},
		},
		{
			name:     "multiple base commands",
			baseCmd:  []string{"chroot", "/host", "iptables"},
			expected: []string{"chroot", "/host", "iptables"},
		},
		{
			name:     "empty base command",
			baseCmd:  []string{},
			expected: []string{},
		},
		{
			name:     "command with flags",
			baseCmd:  []string{"ip", "route"},
			expected: []string{"ip", "route"},
		},
		{
			name:     "nil base command",
			baseCmd:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := newCommand(tt.baseCmd...)
			result := cb.build()
			if !slices.Equal(result, tt.expected) {
				t.Errorf("newCommand(%v).build() = %v, want %v", tt.baseCmd, result, tt.expected)
			}
		})
	}
}

func TestCommandBuilder_Add(t *testing.T) {
	tests := []struct {
		name     string
		baseCmd  []string
		addArgs  [][]string
		expected []string
	}{
		{
			name:     "add single argument",
			baseCmd:  []string{"ls"},
			addArgs:  [][]string{{"-l"}},
			expected: []string{"ls", "-l"},
		},
		{
			name:     "add multiple arguments at once",
			baseCmd:  []string{"ls"},
			addArgs:  [][]string{{"-l", "-a", "-h"}},
			expected: []string{"ls", "-l", "-a", "-h"},
		},
		{
			name:     "add arguments in multiple calls",
			baseCmd:  []string{"ip"},
			addArgs:  [][]string{{"route"}, {"show"}},
			expected: []string{"ip", "route", "show"},
		},
		{
			name:     "add empty arguments",
			baseCmd:  []string{"ls"},
			addArgs:  [][]string{{}},
			expected: []string{"ls"},
		},
		{
			name:     "add to empty base command",
			baseCmd:  []string{},
			addArgs:  [][]string{{"ls", "-l"}},
			expected: []string{"ls", "-l"},
		},
		{
			name:     "chain multiple add calls",
			baseCmd:  []string{"iptables"},
			addArgs:  [][]string{{"-t", "filter"}, {"-L"}, {"INPUT"}},
			expected: []string{"iptables", "-t", "filter", "-L", "INPUT"},
		},
		{
			name:     "add arguments with special characters",
			baseCmd:  []string{"grep"},
			addArgs:  [][]string{{"--include=*.go"}, {"pattern"}},
			expected: []string{"grep", "--include=*.go", "pattern"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := newCommand(tt.baseCmd...)
			for _, args := range tt.addArgs {
				cb.add(args...)
			}
			result := cb.build()
			if !slices.Equal(result, tt.expected) {
				t.Errorf("add() chain result = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCommandBuilder_AddIf(t *testing.T) {
	tests := []struct {
		name      string
		baseCmd   []string
		condition bool
		args      []string
		expected  []string
	}{
		{
			name:      "condition true - adds arguments",
			baseCmd:   []string{"ls"},
			condition: true,
			args:      []string{"-l", "-a"},
			expected:  []string{"ls", "-l", "-a"},
		},
		{
			name:      "condition false - does not add arguments",
			baseCmd:   []string{"ls"},
			condition: false,
			args:      []string{"-l", "-a"},
			expected:  []string{"ls"},
		},
		{
			name:      "condition true with empty args",
			baseCmd:   []string{"ls"},
			condition: true,
			args:      []string{},
			expected:  []string{"ls"},
		},
		{
			name:      "condition false with empty args",
			baseCmd:   []string{"ls"},
			condition: false,
			args:      []string{},
			expected:  []string{"ls"},
		},
		{
			name:      "condition true with single arg",
			baseCmd:   []string{"ip"},
			condition: true,
			args:      []string{"-6"},
			expected:  []string{"ip", "-6"},
		},
		{
			name:      "condition false with single arg",
			baseCmd:   []string{"ip"},
			condition: false,
			args:      []string{"-6"},
			expected:  []string{"ip"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := newCommand(tt.baseCmd...)
			cb.addIf(tt.condition, tt.args...)
			result := cb.build()
			if !slices.Equal(result, tt.expected) {
				t.Errorf("addIf(%v, %v) = %v, want %v", tt.condition, tt.args, result, tt.expected)
			}
		})
	}
}

func TestCommandBuilder_AddIfNotEmpty(t *testing.T) {
	tests := []struct {
		name     string
		baseCmd  []string
		value    string
		args     []string
		expected []string
	}{
		{
			name:     "non-empty value - adds arguments",
			baseCmd:  []string{"iptables"},
			value:    "filter",
			args:     []string{"-t", "filter"},
			expected: []string{"iptables", "-t", "filter"},
		},
		{
			name:     "empty value - does not add arguments",
			baseCmd:  []string{"iptables"},
			value:    "",
			args:     []string{"-t", "filter"},
			expected: []string{"iptables"},
		},
		{
			name:     "whitespace only value - adds arguments",
			baseCmd:  []string{"iptables"},
			value:    "   ",
			args:     []string{"-t", "nat"},
			expected: []string{"iptables", "-t", "nat"},
		},
		{
			name:     "non-empty value with empty args",
			baseCmd:  []string{"ls"},
			value:    "test",
			args:     []string{},
			expected: []string{"ls"},
		},
		{
			name:     "empty value with empty args",
			baseCmd:  []string{"ls"},
			value:    "",
			args:     []string{},
			expected: []string{"ls"},
		},
		{
			name:     "non-empty value with single arg",
			baseCmd:  []string{"ip"},
			value:    "show",
			args:     []string{"route"},
			expected: []string{"ip", "route"},
		},
		{
			name:     "non-empty value with multiple args",
			baseCmd:  []string{"ip"},
			value:    "192.168.1.1",
			args:     []string{"-d", "192.168.1.1"},
			expected: []string{"ip", "-d", "192.168.1.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := newCommand(tt.baseCmd...)
			cb.addIfNotEmpty(tt.value, tt.args...)
			result := cb.build()
			if !slices.Equal(result, tt.expected) {
				t.Errorf("addIfNotEmpty(%q, %v) = %v, want %v", tt.value, tt.args, result, tt.expected)
			}
		})
	}
}

func TestCommandBuilder_Build(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *commandBuilder
		expected []string
	}{
		{
			name: "simple command",
			builder: func() *commandBuilder {
				return newCommand("ls")
			},
			expected: []string{"ls"},
		},
		{
			name: "command built with add",
			builder: func() *commandBuilder {
				return newCommand("ls").add("-l", "-a")
			},
			expected: []string{"ls", "-l", "-a"},
		},
		{
			name: "command built with addIf true",
			builder: func() *commandBuilder {
				return newCommand("ls").addIf(true, "-l")
			},
			expected: []string{"ls", "-l"},
		},
		{
			name: "command built with addIf false",
			builder: func() *commandBuilder {
				return newCommand("ls").addIf(false, "-l")
			},
			expected: []string{"ls"},
		},
		{
			name: "command built with addIfNotEmpty non-empty",
			builder: func() *commandBuilder {
				return newCommand("ls").addIfNotEmpty("value", "-l")
			},
			expected: []string{"ls", "-l"},
		},
		{
			name: "command built with addIfNotEmpty empty",
			builder: func() *commandBuilder {
				return newCommand("ls").addIfNotEmpty("", "-l")
			},
			expected: []string{"ls"},
		},
		{
			name: "empty command",
			builder: func() *commandBuilder {
				return newCommand()
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().build()
			if !slices.Equal(result, tt.expected) {
				t.Errorf("build() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCommandBuilder_ChainedCalls(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *commandBuilder
		expected []string
	}{
		{
			name: "chain add, addIf, addIfNotEmpty",
			builder: func() *commandBuilder {
				return newCommand("iptables").
					add("-t", "filter").
					addIf(true, "-L").
					addIfNotEmpty("INPUT", "INPUT")
			},
			expected: []string{"iptables", "-t", "filter", "-L", "INPUT"},
		},
		{
			name: "chain with mixed conditions",
			builder: func() *commandBuilder {
				return newCommand("ip").
					addIf(true, "route").
					addIf(false, "link").
					add("show").
					addIfNotEmpty("", "default").
					addIfNotEmpty("table", "table", "main")
			},
			expected: []string{"ip", "route", "show", "table", "main"},
		},
		{
			name: "real world iptables example",
			builder: func() *commandBuilder {
				table := "nat"
				command := "-L"
				filterParams := "-n -v"
				return newCommand("iptables").
					addIf(table == "", "-t", "filter").
					addIfNotEmpty(table, "-t", table).
					addIf(command == "", "-L").
					addIfNotEmpty(command, command).
					addIfNotEmpty(filterParams, "-n", "-v")
			},
			expected: []string{"iptables", "-t", "nat", "-L", "-n", "-v"},
		},
		{
			name: "complex chaining with all methods",
			builder: func() *commandBuilder {
				return newCommand("chroot", "/host").
					add("iptables").
					addIf(true, "-t").
					add("filter").
					addIfNotEmpty("INPUT", "-L", "INPUT").
					addIf(false, "-S").
					add("-n")
			},
			expected: []string{"chroot", "/host", "iptables", "-t", "filter", "-L", "INPUT", "-n"},
		},
		{
			name: "multiple build calls return same result",
			builder: func() *commandBuilder {
				cb := newCommand("ls").add("-l")
				cb.build()
				cb.build()
				return cb
			},
			expected: []string{"ls", "-l"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().build()
			if !slices.Equal(result, tt.expected) {
				t.Errorf("chained build() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCommandBuilder_Immutability(t *testing.T) {
	t.Run("build does not modify builder", func(t *testing.T) {
		cb := newCommand("ls").add("-l")
		result1 := cb.build()
		result2 := cb.build()

		if !slices.Equal(result1, result2) {
			t.Errorf("multiple build() calls returned different results: %v vs %v", result1, result2)
		}
	})

	t.Run("modifying returned slice affects builder - no defensive copy", func(t *testing.T) {
		cb := newCommand("ls").add("-l")
		result1 := cb.build()
		result1[0] = "modified"
		result2 := cb.build()

		// Note: The current implementation returns the internal slice directly
		// without making a defensive copy, so modifications affect the builder
		if result2[0] != "modified" {
			t.Errorf("expected modification to affect builder, but it didn't")
		}
	})

	t.Run("continue building after build call", func(t *testing.T) {
		cb := newCommand("ls")
		cb.build()
		cb.add("-l")
		result := cb.build()

		expected := []string{"ls", "-l"}
		if !slices.Equal(result, expected) {
			t.Errorf("build after modification = %v, want %v", result, expected)
		}
	})
}
