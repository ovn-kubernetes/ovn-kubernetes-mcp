package external_tools

import (
	"strings"
	"testing"
)

func TestValidateInterface(t *testing.T) {
	tests := []struct {
		name    string
		iface   string
		wantErr bool
	}{
		{
			name:    "empty interface is valid",
			iface:   "",
			wantErr: false,
		},
		{
			name:    "any interface is valid",
			iface:   "any",
			wantErr: false,
		},
		{
			name:    "valid interface with hyphen",
			iface:   "eth0",
			wantErr: false,
		},
		{
			name:    "valid interface with underscore",
			iface:   "br_ex",
			wantErr: false,
		},
		{
			name:    "valid interface with dot",
			iface:   "eth0.100",
			wantErr: false,
		},
		{
			name:    "valid interface with mixed chars",
			iface:   "veth1a2b-c_d",
			wantErr: false,
		},
		{
			name:    "interface name too long",
			iface:   "verylonginterfacename",
			wantErr: true,
		},
		{
			name:    "interface with space returns error",
			iface:   "eth 0",
			wantErr: true,
		},
		{
			name:    "interface with slash returns error",
			iface:   "eth/0",
			wantErr: true,
		},
		{
			name:    "interface with semicolon returns error",
			iface:   "eth;0",
			wantErr: true,
		},
		{
			name:    "interface starting with hyphen returns error",
			iface:   "-eth0",
			wantErr: true,
		},
		{
			name:    "interface with special chars returns error",
			iface:   "eth@0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInterface(tt.iface)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInterface() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBPFFilter(t *testing.T) {
	tests := []struct {
		name    string
		filter  string
		wantErr bool
	}{
		{
			name:    "empty filter is valid",
			filter:  "",
			wantErr: false,
		},
		{
			name:    "simple tcp filter",
			filter:  "tcp port 80",
			wantErr: false,
		},
		{
			name:    "filter with host and port",
			filter:  "tcp port 80 and host 10.0.0.1",
			wantErr: false,
		},
		{
			name:    "filter with complex expression",
			filter:  "tcp port 80 or udp port 53",
			wantErr: false,
		},
		{
			name:    "filter with parentheses",
			filter:  "(tcp port 80) and (host 10.0.0.1)",
			wantErr: false,
		},
		{
			name:    "filter with semicolon returns error",
			filter:  "tcp port 80; rm -rf /",
			wantErr: true,
		},
		{
			name:    "filter with pipe returns error",
			filter:  "tcp port 80 | cat",
			wantErr: true,
		},
		{
			name:    "filter with ampersand returns error",
			filter:  "tcp port 80 & ls",
			wantErr: true,
		},
		{
			name:    "filter with backtick returns error",
			filter:  "tcp port `echo 80`",
			wantErr: true,
		},
		{
			name:    "filter with dollar returns error",
			filter:  "tcp port $PORT",
			wantErr: true,
		},
		{
			name:    "filter with dollar parentheses returns error",
			filter:  "tcp port $(echo 80)",
			wantErr: true,
		},
		{
			name:    "filter too long returns error",
			filter:  strings.Repeat("a", 1025),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBPFFilter(tt.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBPFFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequireAtLeastNParams(t *testing.T) {
	tests := []struct {
		name     string
		required int
		params   map[string]bool
		wantErr  bool
	}{
		{
			name:     "all params set",
			required: 2,
			params: map[string]bool{
				"duration":     true,
				"packet_count": true,
				"bpf_filter":   true,
			},
			wantErr: false,
		},
		{
			name:     "exactly required params set",
			required: 2,
			params: map[string]bool{
				"duration":     true,
				"packet_count": true,
				"bpf_filter":   false,
			},
			wantErr: false,
		},
		{
			name:     "not enough params set",
			required: 2,
			params: map[string]bool{
				"duration":     true,
				"packet_count": false,
				"bpf_filter":   false,
			},
			wantErr: true,
		},
		{
			name:     "no params set",
			required: 1,
			params: map[string]bool{
				"duration":     false,
				"packet_count": false,
			},
			wantErr: true,
		},
		{
			name:     "require 0 params always succeeds",
			required: 0,
			params: map[string]bool{
				"duration": false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireAtLeastNParams(tt.required, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("requireAtLeastNParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIntMax(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		max       int
		fieldName string
		unit      string
		wantErr   bool
	}{
		{
			name:      "value below max",
			value:     10,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   false,
		},
		{
			name:      "value equals max",
			value:     30,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   false,
		},
		{
			name:      "value exceeds max with unit",
			value:     31,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   true,
		},
		{
			name:      "value exceeds max without unit",
			value:     1001,
			max:       1000,
			fieldName: "packet_count",
			unit:      "",
			wantErr:   true,
		},
		{
			name:      "zero value with max",
			value:     0,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   false,
		},
		{
			name:      "negative value with max",
			value:     -1,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIntMax(tt.value, tt.max, tt.fieldName, tt.unit)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIntMax() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStringWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		defaultValue string
		want         string
	}{
		{
			name:         "empty value returns default",
			value:        "",
			defaultValue: "text",
			want:         "text",
		},
		{
			name:         "non-empty value returns value",
			value:        "pcap",
			defaultValue: "text",
			want:         "pcap",
		},
		{
			name:         "both empty returns empty",
			value:        "",
			defaultValue: "",
			want:         "",
		},
		{
			name:         "whitespace value returns value",
			value:        " ",
			defaultValue: "text",
			want:         " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringWithDefault(tt.value, tt.defaultValue)
			if got != tt.want {
				t.Errorf("stringWithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommandBuilder(t *testing.T) {
	tests := []struct {
		name    string
		builder func() *commandBuilder
		want    []string
	}{
		{
			name: "simple command",
			builder: func() *commandBuilder {
				return newCommand("tcpdump")
			},
			want: []string{"tcpdump"},
		},
		{
			name: "command with args",
			builder: func() *commandBuilder {
				return newCommand("tcpdump").add("-i", "eth0")
			},
			want: []string{"tcpdump", "-i", "eth0"},
		},
		{
			name: "command with conditional args true",
			builder: func() *commandBuilder {
				return newCommand("tcpdump").addIf(true, "-v")
			},
			want: []string{"tcpdump", "-v"},
		},
		{
			name: "command with conditional args false",
			builder: func() *commandBuilder {
				return newCommand("tcpdump").addIf(false, "-v")
			},
			want: []string{"tcpdump"},
		},
		{
			name: "command with addIfNotEmpty with value",
			builder: func() *commandBuilder {
				return newCommand("tcpdump").addIfNotEmpty("tcp port 80", "tcp port 80")
			},
			want: []string{"tcpdump", "tcp port 80"},
		},
		{
			name: "command with addIfNotEmpty without value",
			builder: func() *commandBuilder {
				return newCommand("tcpdump").addIfNotEmpty("", "-f")
			},
			want: []string{"tcpdump"},
		},
		{
			name: "command with chained operations",
			builder: func() *commandBuilder {
				return newCommand("tcpdump").
					add("-i", "eth0").
					addIf(true, "-n").
					addIf(false, "-v").
					addIfNotEmpty("tcp", "tcp").
					addIfNotEmpty("", "-x")
			},
			want: []string{"tcpdump", "-i", "eth0", "-n", "tcp"},
		},
		{
			name: "command with multiple base args",
			builder: func() *commandBuilder {
				return newCommand("timeout", "30s", "retis")
			},
			want: []string{"timeout", "30s", "retis"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.builder().build()
			if len(got) != len(tt.want) {
				t.Errorf("commandBuilder.build() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("commandBuilder.build()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
