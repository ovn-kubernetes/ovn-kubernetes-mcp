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

func TestValidateShellMetacharacters(t *testing.T) {
	tests := []struct {
		name                    string
		param                   string
		shellMetaCharactersType ShellMetaCharactersType
		wantError               bool
	}{
		{
			name:                    "empty string",
			param:                   "",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               false,
		},
		{
			name:                    "valid with hyphens",
			param:                   "--param",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               false,
		},
		{
			name:                    "valid with equals",
			param:                   "key=value",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               false,
		},
		{
			name:                    "valid with spaces",
			param:                   "hello world",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               false,
		},
		{
			name:                    "invalid with semicolon",
			param:                   "test;rm -rf",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with ampersand",
			param:                   "test&background",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with pipe",
			param:                   "test|grep",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with dollar sign",
			param:                   "test$var",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with backtick",
			param:                   "test`whoami`",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with less than",
			param:                   "test<input",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with greater than",
			param:                   "test>output",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with backslash",
			param:                   "test\\escape",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with opening parenthesis",
			param:                   "test(subshell",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with closing parenthesis",
			param:                   "test)subshell",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with multiple metacharacters",
			param:                   "test;ls|grep&",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with semicolon only",
			param:                   ";",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with command substitution",
			param:                   "$(whoami)",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with backtick command substitution",
			param:                   "`id`",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "valid IP address",
			param:                   "192.168.1.1",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               false,
		},
		{
			name:                    "valid CIDR notation",
			param:                   "10.0.0.0/24",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               false,
		},
		{
			name:                    "valid with port number",
			param:                   "127.0.0.1:8080",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               false,
		},
		{
			name:                    "valid with ampersand",
			param:                   "test&background",
			shellMetaCharactersType: ShellMetaCharactersTypeAllowBracketsAllowAmp,
			wantError:               false,
		},
		{
			name:                    "valid with brackets",
			param:                   "test(subshell)",
			shellMetaCharactersType: ShellMetaCharactersTypeAllowBrackets,
			wantError:               false,
		},
		{
			name:                    "invalid with brackets",
			param:                   "test(subshell)",
			shellMetaCharactersType: ShellMetaCharactersTypeDefault,
			wantError:               true,
		},
		{
			name:                    "invalid with dollar sign",
			param:                   "$(subshell)",
			shellMetaCharactersType: ShellMetaCharactersTypeDisallowSpecialCharacters,
			wantError:               true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateShellMetacharacters(tt.param, tt.shellMetaCharactersType)
			if tt.wantError && err == nil {
				t.Errorf("ValidateShellMetacharacters(%q, %s) expected error, got nil", tt.param, tt.shellMetaCharactersType)
			}
			if !tt.wantError && err != nil {
				t.Errorf("ValidateShellMetacharacters(%q, %s) expected no error, got %v", tt.param, tt.shellMetaCharactersType, err)
			}
		})
	}
}
