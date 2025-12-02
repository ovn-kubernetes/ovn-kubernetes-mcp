package kernel

import (
	"testing"
)

func TestValidateIPCommand(t *testing.T) {
	tests := []struct {
		name      string
		ipCommand string
		wantError bool
	}{
		{
			name:      "empty string",
			ipCommand: "",
			wantError: true,
		},
		{
			name:      "numeric string",
			ipCommand: "999",
			wantError: true,
		},
		{
			name:      "valid command - address show",
			ipCommand: "address show",
			wantError: false,
		},
		{
			name:      "valid command - link show",
			ipCommand: "link show",
			wantError: false,
		},
		{
			name:      "valid command - neighbour show",
			ipCommand: "neighbour show",
			wantError: false,
		},
		{
			name:      "valid command - netns show",
			ipCommand: "netns show",
			wantError: false,
		},
		{
			name:      "valid command - route show",
			ipCommand: "route show",
			wantError: false,
		},
		{
			name:      "valid command - rule show",
			ipCommand: "rule show",
			wantError: false,
		},
		{
			name:      "valid command - vrf show",
			ipCommand: "vrf show",
			wantError: false,
		},
		{
			name:      "valid command - addr show",
			ipCommand: "addr show",
			wantError: false,
		},
		{
			name:      "invalid command - neighbor show",
			ipCommand: "neighbor show",
			wantError: true,
		},
		{
			name:      "invalid command - address add",
			ipCommand: "address add",
			wantError: true,
		},
		{
			name:      "invalid command - link set",
			ipCommand: "link set",
			wantError: true,
		},
		{
			name:      "invalid command - route add",
			ipCommand: "route add",
			wantError: true,
		},
		{
			name:      "invalid command - just address",
			ipCommand: "address",
			wantError: true,
		},
		{
			name:      "invalid command - uppercase",
			ipCommand: "ADDRESS SHOW",
			wantError: true,
		},
		{
			name:      "valid command - address show with extra spaces",
			ipCommand: "address  show",
			wantError: false,
		},
		{
			name:      "valid command - with trailing space",
			ipCommand: "address show ",
			wantError: false,
		},
		{
			name:      "valid command - abbreviated with spaces",
			ipCommand: "n s",
			wantError: false,
		},
		{
			name:      "invalid command - l s (can set link)",
			ipCommand: "l s",
			wantError: true,
		},
		{
			name:      "invalid command - xfrm state show",
			ipCommand: "xfrm state show",
			wantError: true,
		},
		{
			name:      "invalid command - xfrm policy show",
			ipCommand: "xfrm policy show",
			wantError: true,
		},
		{
			name:      "valid command - xfrm state list",
			ipCommand: "xfrm state list",
			wantError: false,
		},
		{
			name:      "valid command - xfrm policy list",
			ipCommand: "xfrm policy list",
			wantError: false,
		},
		{
			name:      "valid command - xfrm state list with abbreviated list",
			ipCommand: "xfrm state l",
			wantError: false,
		},
		{
			name:      "valid command - xfrm policy list with abbreviated list",
			ipCommand: "xfrm policy l",
			wantError: false,
		},
		{
			name:      "invalid command - xfrm state get",
			ipCommand: "xfrm state get",
			wantError: true,
		},
		{
			name:      "invalid command - xfrm policy get",
			ipCommand: "xfrm policy get",
			wantError: true,
		},
		{
			name:      "invalid command - xfrm state add",
			ipCommand: "xfrm state add",
			wantError: true,
		},
		{
			name:      "invalid command - xfrm policy add",
			ipCommand: "xfrm policy add",
			wantError: true,
		},
		{
			name:      "invalid command - xfrm state delete",
			ipCommand: "xfrm state delete",
			wantError: true,
		},
		{
			name:      "invalid command - xfrm policy delete",
			ipCommand: "xfrm policy delete",
			wantError: true,
		},
		{
			name:      "invalid command - just xfrm state",
			ipCommand: "xfrm state",
			wantError: true,
		},
		{
			name:      "invalid command - just xfrm policy",
			ipCommand: "xfrm policy",
			wantError: true,
		},
		{
			name:      "invalid command - just xfrm",
			ipCommand: "xfrm show",
			wantError: true,
		},
		{
			name:      "valid command - xfrm state list with trailing spaces",
			ipCommand: "xfrm state list ",
			wantError: false,
		},
		{
			name:      "valid command - xfrm policy list with extra spaces",
			ipCommand: "xfrm  policy  list",
			wantError: false,
		},
		{
			name:      "invalid command - address list",
			ipCommand: "address list",
			wantError: true,
		},
		{
			name:      "invalid command - link list",
			ipCommand: "link list",
			wantError: true,
		},
		{
			name:      "invalid command - route list",
			ipCommand: "route list",
			wantError: true,
		},
		{
			name:      "invalid command - neighbour list",
			ipCommand: "neighbour list",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIPCommand(tt.ipCommand)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateIPCommand(%q) error = %v, wantError %v", tt.ipCommand, err, tt.wantError)
			}
		})
	}
}
