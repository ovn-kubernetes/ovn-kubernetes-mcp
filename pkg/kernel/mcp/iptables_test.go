package kernel

import (
	"testing"
)

func TestValidateTableName(t *testing.T) {
	tests := []struct {
		name      string
		table     string
		wantError bool
	}{
		{
			name:      "empty string",
			table:     "",
			wantError: false,
		},
		{
			name:      "numeric string",
			table:     "123",
			wantError: true,
		},
		{
			name:      "valid table - filter",
			table:     "filter",
			wantError: false,
		},
		{
			name:      "valid table - nat",
			table:     "nat",
			wantError: false,
		},
		{
			name:      "valid table - mangle",
			table:     "mangle",
			wantError: false,
		},
		{
			name:      "valid table - raw",
			table:     "raw",
			wantError: false,
		},
		{
			name:      "valid table - security",
			table:     "security",
			wantError: false,
		},
		{
			name:      "invalid table - random",
			table:     "random",
			wantError: true,
		},
		{
			name:      "invalid table - FILTER uppercase",
			table:     "FILTER",
			wantError: true,
		},
		{
			name:      "invalid table - with spaces",
			table:     "filter ",
			wantError: false,
		},
		{
			name:      "invalid table - custom",
			table:     "custom_table",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTableName(tt.table)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateTableName(%q) error = %v, wantError %v", tt.table, err, tt.wantError)
			}
		})
	}
}

func TestValidateIptablesCommand(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		wantError bool
	}{
		{
			name:      "empty string",
			command:   "",
			wantError: false,
		},
		{
			name:      "numeric string",
			command:   "456",
			wantError: true,
		},
		{
			name:      "valid command - list",
			command:   "-L",
			wantError: false,
		},
		{
			name:      "valid command - list rules",
			command:   "-S",
			wantError: false,
		},
		{
			name:      "invalid command - append",
			command:   "-A",
			wantError: true,
		},
		{
			name:      "invalid command - delete",
			command:   "-D",
			wantError: true,
		},
		{
			name:      "invalid command - insert",
			command:   "-I",
			wantError: true,
		},
		{
			name:      "invalid command - flush",
			command:   "-F",
			wantError: true,
		},
		{
			name:      "invalid command - lowercase l",
			command:   "-l",
			wantError: true,
		},
		{
			name:      "invalid command - with spaces",
			command:   "-L ",
			wantError: false,
		},
		{
			name:      "invalid command - random",
			command:   "list",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIptablesCommand(tt.command)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateIptablesCommand(%q) error = %v, wantError %v", tt.command, err, tt.wantError)
			}
		})
	}
}

func TestIptablesCommand(t *testing.T) {
	tests := []struct {
		name             string
		filterParameters string
		want             string
	}{
		{
			name:             "empty filter parameters",
			filterParameters: "",
			want:             "iptables",
		},
		{
			name:             "ipv6 flag --ipv6",
			filterParameters: "--ipv6",
			want:             "ip6tables",
		},
		{
			name:             "ipv6 short flag -6",
			filterParameters: "-6",
			want:             "ip6tables",
		},
		{
			name:             "ipv6 flag with other parameters",
			filterParameters: "-p tcp --ipv6 --dport 80",
			want:             "ip6tables",
		},
		{
			name:             "ipv6 short flag with other parameters",
			filterParameters: "-p tcp -6 --dport 80",
			want:             "ip6tables",
		},
		{
			name:             "ipv4 parameters only",
			filterParameters: "-p tcp --dport 80",
			want:             "iptables",
		},
		{
			name:             "parameters with number 6 but not ipv6 flag",
			filterParameters: "--dport 6000",
			want:             "iptables",
		},
		{
			name:             "flag containing both - and 6",
			filterParameters: "-p6",
			want:             "ip6tables",
		},
		{
			name:             "flag containing both - and 6 in middle",
			filterParameters: "-m6state",
			want:             "ip6tables",
		},
		{
			name:             "multiple flags with ipv6 at end",
			filterParameters: "-p tcp -m state --state NEW --ipv6",
			want:             "ip6tables",
		},
		{
			name:             "multiple flags with -6 at beginning",
			filterParameters: "-6 -p tcp -m state --state NEW",
			want:             "ip6tables",
		},
		{
			name:             "protocol number without dash",
			filterParameters: "-p tcp --sport 6379",
			want:             "iptables",
		},
		{
			name:             "dash without 6",
			filterParameters: "-p tcp -m state",
			want:             "iptables",
		},
		{
			name:             "number 6 without dash",
			filterParameters: "6",
			want:             "iptables",
		},
		{
			name:             "chain name with 6",
			filterParameters: "-A INPUT6",
			want:             "iptables",
		},
		{
			name:             "ipv6 address in parameter",
			filterParameters: "-d 2001:db8::1",
			want:             "iptables",
		},
		{
			name:             "combined short flags with 6",
			filterParameters: "-t6n",
			want:             "ip6tables",
		},
		{
			name:             "whitespace variations with ipv6",
			filterParameters: "  --ipv6  -p tcp  ",
			want:             "ip6tables",
		},
		{
			name:             "whitespace variations without ipv6",
			filterParameters: "  -p tcp  --dport 443  ",
			want:             "iptables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := iptablesCommand(tt.filterParameters)
			if got != tt.want {
				t.Errorf("iptablesCommand(%q) = %v, want %v", tt.filterParameters, got, tt.want)
			}
		})
	}
}
