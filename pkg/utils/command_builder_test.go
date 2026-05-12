package utils

import (
	"slices"
	"testing"
)

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
			cb := NewCommand(tt.baseCmd...)
			result := cb.Build()
			if !slices.Equal(result, tt.expected) {
				t.Errorf("NewCommand(%v).Build() = %v, want %v", tt.baseCmd, result, tt.expected)
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
			cb := NewCommand(tt.baseCmd...)
			for _, args := range tt.addArgs {
				cb.Add(args...)
			}
			result := cb.Build()
			if !slices.Equal(result, tt.expected) {
				t.Errorf("Add() chain result = %v, want %v", result, tt.expected)
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
			cb := NewCommand(tt.baseCmd...)
			cb.AddIf(tt.condition, tt.args...)
			result := cb.Build()
			if !slices.Equal(result, tt.expected) {
				t.Errorf("AddIf(%v, %v) = %v, want %v", tt.condition, tt.args, result, tt.expected)
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
			cb := NewCommand(tt.baseCmd...)
			cb.AddIfNotEmpty(tt.value, tt.args...)
			result := cb.Build()
			if !slices.Equal(result, tt.expected) {
				t.Errorf("AddIfNotEmpty(%q, %v) = %v, want %v", tt.value, tt.args, result, tt.expected)
			}
		})
	}
}

func TestCommandBuilder_Build(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *CommandBuilder
		expected []string
	}{
		{
			name: "simple command",
			builder: func() *CommandBuilder {
				return NewCommand("ls")
			},
			expected: []string{"ls"},
		},
		{
			name: "command built with add",
			builder: func() *CommandBuilder {
				return NewCommand("ls").Add("-l", "-a")
			},
			expected: []string{"ls", "-l", "-a"},
		},
		{
			name: "command built with addIf true",
			builder: func() *CommandBuilder {
				return NewCommand("ls").AddIf(true, "-l")
			},
			expected: []string{"ls", "-l"},
		},
		{
			name: "command built with addIf false",
			builder: func() *CommandBuilder {
				return NewCommand("ls").AddIf(false, "-l")
			},
			expected: []string{"ls"},
		},
		{
			name: "command built with addIfNotEmpty non-empty",
			builder: func() *CommandBuilder {
				return NewCommand("ls").AddIfNotEmpty("value", "-l")
			},
			expected: []string{"ls", "-l"},
		},
		{
			name: "command built with addIfNotEmpty empty",
			builder: func() *CommandBuilder {
				return NewCommand("ls").AddIfNotEmpty("", "-l")
			},
			expected: []string{"ls"},
		},
		{
			name: "empty command",
			builder: func() *CommandBuilder {
				return NewCommand()
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			if !slices.Equal(result, tt.expected) {
				t.Errorf("build() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCommandBuilder_ChainedCalls(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *CommandBuilder
		expected []string
	}{
		{
			name: "chain add, addIf, addIfNotEmpty",
			builder: func() *CommandBuilder {
				return NewCommand("iptables").
					Add("-t", "filter").
					AddIf(true, "-L").
					AddIfNotEmpty("INPUT", "INPUT")
			},
			expected: []string{"iptables", "-t", "filter", "-L", "INPUT"},
		},
		{
			name: "chain with mixed conditions",
			builder: func() *CommandBuilder {
				return NewCommand("ip").
					AddIf(true, "route").
					AddIf(false, "link").
					Add("show").
					AddIfNotEmpty("", "default").
					AddIfNotEmpty("table", "table", "main")
			},
			expected: []string{"ip", "route", "show", "table", "main"},
		},
		{
			name: "real world iptables example",
			builder: func() *CommandBuilder {
				table := "nat"
				command := "-L"
				filterParams := "-n -v"
				return NewCommand("iptables").
					AddIf(table == "", "-t", "filter").
					AddIfNotEmpty(table, "-t", table).
					AddIf(command == "", "-L").
					AddIfNotEmpty(command, command).
					AddIfNotEmpty(filterParams, "-n", "-v")
			},
			expected: []string{"iptables", "-t", "nat", "-L", "-n", "-v"},
		},
		{
			name: "complex chaining with all methods",
			builder: func() *CommandBuilder {
				return NewCommand("chroot", "/host").
					Add("iptables").
					AddIf(true, "-t").
					Add("filter").
					AddIfNotEmpty("INPUT", "-L", "INPUT").
					AddIf(false, "-S").
					Add("-n")
			},
			expected: []string{"chroot", "/host", "iptables", "-t", "filter", "-L", "INPUT", "-n"},
		},
		{
			name: "multiple build calls return same result",
			builder: func() *CommandBuilder {
				cb := NewCommand("ls").Add("-l")
				cb.Build()
				cb.Build()
				return cb
			},
			expected: []string{"ls", "-l"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			if !slices.Equal(result, tt.expected) {
				t.Errorf("chained build() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCommandBuilder_Immutability(t *testing.T) {
	t.Run("build does not modify builder", func(t *testing.T) {
		cb := NewCommand("ls").Add("-l")
		result1 := cb.Build()
		result2 := cb.Build()

		if !slices.Equal(result1, result2) {
			t.Errorf("multiple build() calls returned different results: %v vs %v", result1, result2)
		}
	})

	t.Run("modifying returned slice does not affect builder", func(t *testing.T) {
		cb := NewCommand("ls").Add("-l")
		result1 := cb.Build()
		result1[0] = "modified"
		result2 := cb.Build()

		// Note: The current implementation returns a cloned slice and not the
		// internal slice, so modifications do not affect the builder
		if result2[0] != "ls" {
			t.Errorf("expected modification to not affect builder, but it did")
		}
	})

	t.Run("continue building after build call", func(t *testing.T) {
		cb := NewCommand("ls")
		cb.Build()
		cb.Add("-l")
		result := cb.Build()

		expected := []string{"ls", "-l"}
		if !slices.Equal(result, expected) {
			t.Errorf("build after modification = %v, want %v", result, expected)
		}
	})
}
