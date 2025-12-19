package external_tools

import (
	"testing"
)

func TestValidatePwruCommand(t *testing.T) {
	tests := []struct {
		name             string
		filter           string
		outputLimitLines int
		wantError        bool
	}{
		// Filter validation tests
		{
			name:             "empty filter",
			filter:           "",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "simple tcp filter",
			filter:           "tcp port 80",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "simple udp filter",
			filter:           "udp port 53",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "simple icmp filter",
			filter:           "icmp",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with host",
			filter:           "host 192.168.1.1",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with host and port",
			filter:           "host 10.0.0.1 and port 80",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with src and dst",
			filter:           "src 10.0.0.1 and dst 10.0.0.2",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "complex filter with or",
			filter:           "tcp port 80 or udp port 53",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "complex filter with and",
			filter:           "tcp and port 443 and host 10.0.0.1",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with parentheses",
			filter:           "(tcp port 80) and (host 10.0.0.1)",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with not",
			filter:           "not port 22",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with port range",
			filter:           "portrange 8000-9000",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with greater/less operators",
			filter:           "greater 100",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with net",
			filter:           "net 192.168.0.0/24",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with vlan",
			filter:           "vlan 100",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with ip proto",
			filter:           "ip proto 6",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with ether",
			filter:           "ether host ff:ff:ff:ff:ff:ff",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with broadcast",
			filter:           "broadcast",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "filter with multicast",
			filter:           "multicast",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "complex nested filter",
			filter:           "((tcp and port 80) or (udp and port 53)) and host 10.0.0.1",
			outputLimitLines: 100,
			wantError:        false,
		},
		{
			name:             "invalid filter - semicolon injection",
			filter:           "tcp port 80; rm -rf /",
			outputLimitLines: 100,
			wantError:        true,
		},
		{
			name:             "invalid filter - pipe injection",
			filter:           "tcp port 80 | cat",
			outputLimitLines: 100,
			wantError:        true,
		},
		{
			name:             "invalid filter - ampersand injection",
			filter:           "tcp port 80 & ls",
			outputLimitLines: 100,
			wantError:        true,
		},
		{
			name:             "invalid filter - backtick injection",
			filter:           "tcp port `echo 80`",
			outputLimitLines: 100,
			wantError:        true,
		},
		{
			name:             "invalid filter - dollar sign injection",
			filter:           "tcp port $PORT",
			outputLimitLines: 100,
			wantError:        true,
		},
		{
			name:             "invalid filter - command substitution",
			filter:           "tcp port $(echo 80)",
			outputLimitLines: 100,
			wantError:        true,
		},
		{
			name:             "invalid filter - newline injection",
			filter:           "tcp port 80\nrm -rf /",
			outputLimitLines: 100,
			wantError:        true,
		},
		{
			name:             "invalid filter - null byte injection",
			filter:           "tcp port 80\x00rm -rf /",
			outputLimitLines: 100,
			wantError:        true,
		},
		// Output limit lines validation tests
		{
			name:             "zero output limit lines - should use default",
			filter:           "",
			outputLimitLines: 0,
			wantError:        false,
		},
		{
			name:             "negative output limit lines",
			filter:           "",
			outputLimitLines: -1,
			wantError:        true,
		},
		{
			name:             "minimum valid output limit lines",
			filter:           "",
			outputLimitLines: 1,
			wantError:        false,
		},
		{
			name:             "output limit lines within limit",
			filter:           "",
			outputLimitLines: 500,
			wantError:        false,
		},
		{
			name:             "output limit lines at max",
			filter:           "",
			outputLimitLines: MaxOutputLimitLines,
			wantError:        false,
		},
		{
			name:             "output limit lines just over max",
			filter:           "",
			outputLimitLines: MaxOutputLimitLines + 1,
			wantError:        true,
		},
		{
			name:             "output limit lines exceeds max",
			filter:           "",
			outputLimitLines: 1500,
			wantError:        true,
		},
		{
			name:             "very large output limit lines",
			filter:           "",
			outputLimitLines: 999999,
			wantError:        true,
		},
		// Combined validation tests
		{
			name:             "valid filter and valid output limit",
			filter:           "tcp port 80",
			outputLimitLines: 500,
			wantError:        false,
		},
		{
			name:             "valid filter but invalid output limit",
			filter:           "tcp port 80",
			outputLimitLines: 2000,
			wantError:        true,
		},
		{
			name:             "invalid filter but valid output limit",
			filter:           "tcp port 80; ls",
			outputLimitLines: 500,
			wantError:        true,
		},
		{
			name:             "invalid filter and invalid output limit",
			filter:           "tcp port 80; ls",
			outputLimitLines: 2000,
			wantError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate filter
			err := ValidateBPFFilter(tt.filter)
			if err != nil {
				if !tt.wantError {
					t.Errorf("ValidateBPFFilter(%q) error = %v, wantError %v", tt.filter, err, tt.wantError)
				}
				return
			}

			// Validate output limit lines
			outputLimitLines := tt.outputLimitLines
			if outputLimitLines == 0 {
				outputLimitLines = MaxOutputLimitLines
			}
			err = validateIntMax(outputLimitLines, MaxOutputLimitLines, "output_limit_lines", "")
			if (err != nil) != tt.wantError {
				t.Errorf("validateIntMax(%d) error = %v, wantError %v", tt.outputLimitLines, err, tt.wantError)
			}
		})
	}
}
