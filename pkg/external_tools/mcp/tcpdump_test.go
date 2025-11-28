package external_tools

import (
	"testing"
)

func TestTcpdumpValidation(t *testing.T) {
	tests := []struct {
		name      string
		iface     string
		bpfFilter string
		wantErr   bool
	}{
		{
			name:      "valid interface and filter",
			iface:     "eth0",
			bpfFilter: "tcp port 80",
			wantErr:   false,
		},
		{
			name:      "valid any interface with filter",
			iface:     "any",
			bpfFilter: "tcp port 80",
			wantErr:   false,
		},
		{
			name:      "invalid interface with special chars",
			iface:     "eth;0",
			bpfFilter: "tcp port 80",
			wantErr:   true,
		},
		{
			name:      "valid interface with invalid filter",
			iface:     "eth0",
			bpfFilter: "tcp port 80; rm -rf /",
			wantErr:   true,
		},
		{
			name:      "interface too long",
			iface:     "verylonginterfacename",
			bpfFilter: "tcp port 80",
			wantErr:   true,
		},
		{
			name:      "filter with command injection",
			iface:     "eth0",
			bpfFilter: "tcp port $(echo 80)",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ifaceErr := ValidateInterface(tt.iface)
			filterErr := ValidateBPFFilter(tt.bpfFilter)

			gotErr := ifaceErr != nil || filterErr != nil
			if gotErr != tt.wantErr {
				t.Errorf("Validation errors: interface=%v, filter=%v, wantErr=%v", ifaceErr, filterErr, tt.wantErr)
			}
		})
	}
}

func TestTcpdumpConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{
			name:     "MaxTcpdumpDuration is 30",
			constant: MaxTcpdumpDuration,
			expected: 30,
		},
		{
			name:     "MaxPacketCount is 1000",
			constant: MaxPacketCount,
			expected: 1000,
		},
		{
			name:     "DefaultSnaplen is 96",
			constant: DefaultSnaplen,
			expected: 96,
		},
		{
			name:     "MaxSnaplen is 262",
			constant: MaxSnaplen,
			expected: 262,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Constant value = %v, want %v", tt.constant, tt.expected)
			}
		})
	}
}
