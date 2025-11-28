package nettools

import (
	"slices"
	"strings"
	"testing"
)

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
			name:      "zero value",
			value:     0,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   false,
		},
		{
			name:      "minimum valid value",
			value:     1,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   false,
		},
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
			name:      "value just over max",
			value:     31,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   true,
		},
		{
			name:      "value exceeds max with unit",
			value:     100,
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
			name:      "negative value -1",
			value:     -1,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   true,
		},
		{
			name:      "large negative value",
			value:     -100,
			max:       30,
			fieldName: "duration",
			unit:      "seconds",
			wantErr:   true,
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

func TestValidateInterface(t *testing.T) {
	tests := []struct {
		name      string
		iface     string
		wantError bool
	}{
		// Valid interface names
		{
			name:      "empty interface",
			iface:     "",
			wantError: false,
		},
		{
			name:      "any interface",
			iface:     "any",
			wantError: false,
		},
		{
			name:      "valid eth0",
			iface:     "eth0",
			wantError: false,
		},
		{
			name:      "valid interface with dot",
			iface:     "eth0.100",
			wantError: false,
		},
		{
			name:      "valid interface with underscore",
			iface:     "br_ex",
			wantError: false,
		},
		{
			name:      "valid interface with hyphen",
			iface:     "veth-abc",
			wantError: false,
		},
		{
			name:      "valid interface wlan0",
			iface:     "wlan0",
			wantError: false,
		},
		{
			name:      "valid interface lo",
			iface:     "lo",
			wantError: false,
		},
		{
			name:      "valid interface br-int",
			iface:     "br-int",
			wantError: false,
		},
		{
			name:      "valid interface with numbers",
			iface:     "veth123abc",
			wantError: false,
		},
		{
			name:      "valid 15 character interface",
			iface:     "123456789012345",
			wantError: false,
		},
		// Invalid interface names
		{
			name:      "interface too long (16 chars)",
			iface:     "1234567890123456",
			wantError: true,
		},
		{
			name:      "interface too long",
			iface:     "verylonginterfacename",
			wantError: true,
		},
		{
			name:      "interface with semicolon",
			iface:     "eth;0",
			wantError: true,
		},
		{
			name:      "interface with space",
			iface:     "eth 0",
			wantError: true,
		},
		{
			name:      "interface starting with hyphen",
			iface:     "-eth0",
			wantError: true,
		},
		{
			name:      "interface starting with dot",
			iface:     ".eth0",
			wantError: true,
		},
		{
			name:      "interface starting with underscore",
			iface:     "_eth0",
			wantError: true,
		},
		{
			name:      "interface with pipe",
			iface:     "eth|0",
			wantError: true,
		},
		{
			name:      "interface with ampersand",
			iface:     "eth&0",
			wantError: true,
		},
		{
			name:      "interface with backtick",
			iface:     "eth`0",
			wantError: true,
		},
		{
			name:      "interface with dollar sign",
			iface:     "eth$0",
			wantError: true,
		},
		{
			name:      "interface with newline",
			iface:     "eth0\n",
			wantError: true,
		},
		{
			name:      "interface with parentheses",
			iface:     "eth(0)",
			wantError: true,
		},
		{
			name:      "interface with brackets",
			iface:     "eth[0]",
			wantError: true,
		},
		{
			name:      "interface with slash",
			iface:     "eth/0",
			wantError: true,
		},
		{
			name:      "interface with backslash",
			iface:     "eth\\0",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInterface(tt.iface)
			if (err != nil) != tt.wantError {
				t.Errorf("validateInterface(%q) error = %v, wantError %v", tt.iface, err, tt.wantError)
			}
		})
	}
}

func TestValidatePacketFilter(t *testing.T) {
	tests := []struct {
		name      string
		filter    string
		wantError bool
	}{
		// Valid filters
		{
			name:      "empty filter",
			filter:    "",
			wantError: false,
		},
		{
			name:      "simple tcp filter",
			filter:    "tcp port 80",
			wantError: false,
		},
		{
			name:      "simple udp filter",
			filter:    "udp port 53",
			wantError: false,
		},
		{
			name:      "simple icmp filter",
			filter:    "icmp",
			wantError: false,
		},
		{
			name:      "filter with host",
			filter:    "host 192.168.1.1",
			wantError: false,
		},
		{
			name:      "filter with host and port",
			filter:    "host 10.0.0.1 and port 80",
			wantError: false,
		},
		{
			name:      "filter with src and dst",
			filter:    "src 10.0.0.1 and dst 10.0.0.2",
			wantError: false,
		},
		{
			name:      "complex filter with or",
			filter:    "tcp port 80 or udp port 53",
			wantError: false,
		},
		{
			name:      "complex filter with and",
			filter:    "tcp and port 443 and host 10.0.0.1",
			wantError: false,
		},
		{
			name:      "filter with parentheses",
			filter:    "(tcp port 80) and (host 10.0.0.1)",
			wantError: false,
		},
		{
			name:      "filter with not",
			filter:    "not port 22",
			wantError: false,
		},
		{
			name:      "filter with port range",
			filter:    "portrange 8000-9000",
			wantError: false,
		},
		{
			name:      "filter with greater/less operators",
			filter:    "greater 100",
			wantError: false,
		},
		{
			name:      "filter with net",
			filter:    "net 192.168.0.0/24",
			wantError: false,
		},
		{
			name:      "filter with vlan",
			filter:    "vlan 100",
			wantError: false,
		},
		{
			name:      "filter with ip proto",
			filter:    "ip proto 6",
			wantError: false,
		},
		{
			name:      "filter with ether",
			filter:    "ether host ff:ff:ff:ff:ff:ff",
			wantError: false,
		},
		{
			name:      "filter with broadcast",
			filter:    "broadcast",
			wantError: false,
		},
		{
			name:      "filter with multicast",
			filter:    "multicast",
			wantError: false,
		},
		{
			name:      "complex nested filter",
			filter:    "((tcp and port 80) or (udp and port 53)) and host 10.0.0.1",
			wantError: false,
		},
		{
			name:      "filter with IPv6",
			filter:    "ip6 and host 2001:db8::1",
			wantError: false,
		},
		{
			name:      "filter with less",
			filter:    "less 100",
			wantError: false,
		},
		// Invalid filters - command injection attempts
		{
			name:      "invalid filter - semicolon injection",
			filter:    "tcp port 80; rm -rf /",
			wantError: true,
		},
		{
			name:      "invalid filter - pipe injection",
			filter:    "tcp port 80 | cat",
			wantError: true,
		},
		{
			name:      "invalid filter - ampersand injection",
			filter:    "tcp port 80 & ls",
			wantError: true,
		},
		{
			name:      "invalid filter - backtick injection",
			filter:    "tcp port `echo 80`",
			wantError: true,
		},
		{
			name:      "invalid filter - dollar sign injection",
			filter:    "tcp port $PORT",
			wantError: true,
		},
		{
			name:      "invalid filter - command substitution",
			filter:    "tcp port $(echo 80)",
			wantError: true,
		},
		{
			name:      "invalid filter - newline injection",
			filter:    "tcp port 80\nrm -rf /",
			wantError: true,
		},
		{
			name:      "invalid filter - null byte injection",
			filter:    "tcp port 80\x00rm -rf /",
			wantError: true,
		},
		{
			name:      "invalid filter - double pipe",
			filter:    "tcp port 80 || ls",
			wantError: true,
		},
		{
			name:      "invalid filter - double ampersand",
			filter:    "tcp port 80 && ls",
			wantError: true,
		},
		// Invalid filters - length validation
		{
			name:      "filter at max length (1024 chars)",
			filter:    strings.Repeat("a", 1024),
			wantError: false,
		},
		{
			name:      "filter exceeds max length (1025 chars)",
			filter:    strings.Repeat("a", 1025),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePacketFilter(tt.filter)
			if (err != nil) != tt.wantError {
				t.Errorf("validatePacketFilter(%q) error = %v, wantError %v", tt.filter, err, tt.wantError)
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
			if !slices.Equal(got, tt.want) {
				t.Errorf("commandBuilder.build() = %v, want %v", got, tt.want)
			}
		})
	}
}
