package kernel

import (
	"testing"
)

func TestValidateNFTCommand(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		wantError bool
	}{
		{
			name:      "empty string",
			command:   "",
			wantError: true,
		},
		{
			name:      "numeric string",
			command:   "789",
			wantError: true,
		},
		{
			name:      "valid command - list ruleset",
			command:   "list ruleset",
			wantError: false,
		},
		{
			name:      "valid command - list tables",
			command:   "list tables",
			wantError: false,
		},
		{
			name:      "valid command - list chains",
			command:   "list chains",
			wantError: false,
		},
		{
			name:      "valid command - list sets",
			command:   "list sets",
			wantError: false,
		},
		{
			name:      "valid command - list maps",
			command:   "list maps",
			wantError: false,
		},
		{
			name:      "valid command - list flowtables",
			command:   "list flowtables",
			wantError: false,
		},
		{
			name:      "invalid command - add",
			command:   "add table",
			wantError: true,
		},
		{
			name:      "invalid command - delete",
			command:   "delete table",
			wantError: true,
		},
		{
			name:      "invalid command - flush",
			command:   "flush ruleset",
			wantError: true,
		},
		{
			name:      "invalid command - list only",
			command:   "list",
			wantError: true,
		},
		{
			name:      "invalid command - extra spaces",
			command:   "list  ruleset",
			wantError: true,
		},
		{
			name:      "invalid command - uppercase",
			command:   "LIST RULESET",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNFTCommand(tt.command)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateNFTCommand(%q) error = %v, wantError %v", tt.command, err, tt.wantError)
			}
		})
	}
}

func TestValidateNFTAddressFamily(t *testing.T) {
	tests := []struct {
		name       string
		addrFamily string
		wantError  bool
	}{
		{
			name:       "empty string",
			addrFamily: "",
			wantError:  false,
		},
		{
			name:       "numeric string",
			addrFamily: "101",
			wantError:  true,
		},
		{
			name:       "valid family - ip",
			addrFamily: "ip",
			wantError:  false,
		},
		{
			name:       "valid family - ip6",
			addrFamily: "ip6",
			wantError:  false,
		},
		{
			name:       "valid family - inet",
			addrFamily: "inet",
			wantError:  false,
		},
		{
			name:       "valid family - arp",
			addrFamily: "arp",
			wantError:  false,
		},
		{
			name:       "valid family - bridge",
			addrFamily: "bridge",
			wantError:  false,
		},
		{
			name:       "valid family - netdev",
			addrFamily: "netdev",
			wantError:  false,
		},
		{
			name:       "invalid family - ipv4",
			addrFamily: "ipv4",
			wantError:  true,
		},
		{
			name:       "invalid family - ipv6",
			addrFamily: "ipv6",
			wantError:  true,
		},
		{
			name:       "invalid family - uppercase",
			addrFamily: "IP",
			wantError:  true,
		},
		{
			name:       "invalid family - with spaces",
			addrFamily: "ip ",
			wantError:  false,
		},
		{
			name:       "invalid family - random",
			addrFamily: "random",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNFTAddressFamily(tt.addrFamily)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateNFTAddressFamily(%q) error = %v, wantError %v", tt.addrFamily, err, tt.wantError)
			}
		})
	}
}
