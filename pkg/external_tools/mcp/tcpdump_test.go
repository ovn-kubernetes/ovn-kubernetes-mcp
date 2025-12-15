package external_tools

import (
	"testing"
)

func TestValidateTcpdumpCommand(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		filter      string
		packetCount int
		snaplen     int
		wantError   bool
	}{
		// Interface validation tests
		{
			name:        "empty interface",
			iface:       "",
			filter:      "",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "any interface",
			iface:       "any",
			filter:      "",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "valid eth0",
			iface:       "eth0",
			filter:      "",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "valid interface with dot",
			iface:       "eth0.100",
			filter:      "",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "valid interface with underscore",
			iface:       "br_ex",
			filter:      "",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "valid interface with hyphen",
			iface:       "veth-abc",
			filter:      "",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "interface too long",
			iface:       "verylonginterfacename",
			filter:      "",
			packetCount: 100,
			snaplen:     96,
			wantError:   true,
		},
		{
			name:        "interface with semicolon",
			iface:       "eth;0",
			filter:      "",
			packetCount: 100,
			snaplen:     96,
			wantError:   true,
		},
		{
			name:        "interface with space",
			iface:       "eth 0",
			filter:      "",
			packetCount: 100,
			snaplen:     96,
			wantError:   true,
		},
		{
			name:        "interface starting with hyphen",
			iface:       "-eth0",
			filter:      "",
			packetCount: 100,
			snaplen:     96,
			wantError:   true,
		},
		// BPF Filter validation tests
		{
			name:        "empty filter",
			iface:       "",
			filter:      "",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "simple tcp filter",
			iface:       "",
			filter:      "tcp port 80",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "simple udp filter",
			iface:       "",
			filter:      "udp port 53",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "simple icmp filter",
			iface:       "",
			filter:      "icmp",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "filter with host",
			iface:       "",
			filter:      "host 192.168.1.1",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "filter with host and port",
			iface:       "",
			filter:      "host 10.0.0.1 and port 80",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "complex filter with or",
			iface:       "",
			filter:      "tcp port 80 or udp port 53",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "filter with parentheses",
			iface:       "",
			filter:      "(tcp port 80) and (host 10.0.0.1)",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "filter with not",
			iface:       "",
			filter:      "not port 22",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "filter with port range",
			iface:       "",
			filter:      "portrange 8000-9000",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "filter with net",
			iface:       "",
			filter:      "net 192.168.0.0/24",
			packetCount: 100,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "invalid filter - semicolon injection",
			iface:       "",
			filter:      "tcp port 80; rm -rf /",
			packetCount: 100,
			snaplen:     96,
			wantError:   true,
		},
		{
			name:        "invalid filter - pipe injection",
			iface:       "",
			filter:      "tcp port 80 | cat",
			packetCount: 100,
			snaplen:     96,
			wantError:   true,
		},
		{
			name:        "invalid filter - ampersand injection",
			iface:       "",
			filter:      "tcp port 80 & ls",
			packetCount: 100,
			snaplen:     96,
			wantError:   true,
		},
		{
			name:        "invalid filter - backtick injection",
			iface:       "",
			filter:      "tcp port `echo 80`",
			packetCount: 100,
			snaplen:     96,
			wantError:   true,
		},
		{
			name:        "invalid filter - dollar sign injection",
			iface:       "",
			filter:      "tcp port $PORT",
			packetCount: 100,
			snaplen:     96,
			wantError:   true,
		},
		{
			name:        "invalid filter - command substitution",
			iface:       "",
			filter:      "tcp port $(echo 80)",
			packetCount: 100,
			snaplen:     96,
			wantError:   true,
		},
		{
			name:        "invalid filter - newline injection",
			iface:       "",
			filter:      "tcp port 80\nrm -rf /",
			packetCount: 100,
			snaplen:     96,
			wantError:   true,
		},
		// Packet count validation tests
		{
			name:        "zero packet count - should use default",
			iface:       "",
			filter:      "",
			packetCount: 0,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "negative packet count",
			iface:       "",
			filter:      "",
			packetCount: -1,
			snaplen:     96,
			wantError:   true,
		},
		{
			name:        "minimum valid packet count",
			iface:       "",
			filter:      "",
			packetCount: 1,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "packet count within limit",
			iface:       "",
			filter:      "",
			packetCount: 500,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "packet count at max",
			iface:       "",
			filter:      "",
			packetCount: MaxPacketCount,
			snaplen:     96,
			wantError:   false,
		},
		{
			name:        "packet count just over max",
			iface:       "",
			filter:      "",
			packetCount: MaxPacketCount + 1,
			snaplen:     96,
			wantError:   true,
		},
		{
			name:        "packet count exceeds max",
			iface:       "",
			filter:      "",
			packetCount: 1500,
			snaplen:     96,
			wantError:   true,
		},
		// Snaplen validation tests
		{
			name:        "zero snaplen - should use default",
			iface:       "",
			filter:      "",
			packetCount: 100,
			snaplen:     0,
			wantError:   false,
		},
		{
			name:        "negative snaplen",
			iface:       "",
			filter:      "",
			packetCount: 100,
			snaplen:     -1,
			wantError:   true,
		},
		{
			name:        "minimum valid snaplen",
			iface:       "",
			filter:      "",
			packetCount: 100,
			snaplen:     1,
			wantError:   false,
		},
		{
			name:        "snaplen within limit",
			iface:       "",
			filter:      "",
			packetCount: 100,
			snaplen:     128,
			wantError:   false,
		},
		{
			name:        "snaplen at max",
			iface:       "",
			filter:      "",
			packetCount: 100,
			snaplen:     MaxSnaplen,
			wantError:   false,
		},
		{
			name:        "snaplen just over max",
			iface:       "",
			filter:      "",
			packetCount: 100,
			snaplen:     MaxSnaplen + 1,
			wantError:   true,
		},
		{
			name:        "snaplen exceeds max",
			iface:       "",
			filter:      "",
			packetCount: 100,
			snaplen:     300,
			wantError:   true,
		},
		// Combined validation tests
		{
			name:        "all valid parameters",
			iface:       "eth0",
			filter:      "tcp port 80",
			packetCount: 500,
			snaplen:     128,
			wantError:   false,
		},
		{
			name:        "valid interface but invalid filter",
			iface:       "eth0",
			filter:      "tcp port 80; ls",
			packetCount: 500,
			snaplen:     128,
			wantError:   true,
		},
		{
			name:        "invalid interface but valid filter",
			iface:       "eth;0",
			filter:      "tcp port 80",
			packetCount: 500,
			snaplen:     128,
			wantError:   true,
		},
		{
			name:        "valid interface and filter but invalid packet count",
			iface:       "eth0",
			filter:      "tcp port 80",
			packetCount: 2000,
			snaplen:     128,
			wantError:   true,
		},
		{
			name:        "valid interface and filter but invalid snaplen",
			iface:       "eth0",
			filter:      "tcp port 80",
			packetCount: 500,
			snaplen:     300,
			wantError:   true,
		},
		{
			name:        "all parameters invalid",
			iface:       "eth;0",
			filter:      "tcp port 80; ls",
			packetCount: 2000,
			snaplen:     300,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate interface
			err := ValidateInterface(tt.iface)
			if err != nil {
				if !tt.wantError {
					t.Errorf("ValidateInterface(%q) error = %v, wantError %v", tt.iface, err, tt.wantError)
				}
				return
			}

			// Validate BPF filter
			err = ValidateBPFFilter(tt.filter)
			if err != nil {
				if !tt.wantError {
					t.Errorf("ValidateBPFFilter(%q) error = %v, wantError %v", tt.filter, err, tt.wantError)
				}
				return
			}

			// Validate packet count
			packetCount := tt.packetCount
			if packetCount == 0 {
				packetCount = MaxPacketCount
			}
			err = validateIntMax(packetCount, MaxPacketCount, "packet_count", "")
			if err != nil {
				if !tt.wantError {
					t.Errorf("validateIntMax(packetCount=%d) error = %v, wantError %v", tt.packetCount, err, tt.wantError)
				}
				return
			}

			// Validate snaplen
			snaplen := tt.snaplen
			if snaplen == 0 {
				snaplen = DefaultSnaplen
			}
			err = validateIntMax(snaplen, MaxSnaplen, "snaplen", "bytes")
			if (err != nil) != tt.wantError {
				t.Errorf("validateIntMax(snaplen=%d) error = %v, wantError %v", tt.snaplen, err, tt.wantError)
			}
		})
	}
}
