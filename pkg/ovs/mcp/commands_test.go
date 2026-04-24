package mcp

import (
	"testing"
)

func TestValidateBridgeName(t *testing.T) {
	tests := []struct {
		name    string
		bridge  string
		wantErr bool
	}{
		{
			name:    "valid bridge name with hyphen",
			bridge:  "br-int",
			wantErr: false,
		},
		{
			name:    "valid bridge name with underscore",
			bridge:  "br_ex",
			wantErr: false,
		},
		{
			name:    "valid bridge name alphanumeric",
			bridge:  "br0",
			wantErr: false,
		},
		{
			name:    "valid bridge name mixed",
			bridge:  "br-local_123",
			wantErr: false,
		},
		{
			name:    "empty bridge name returns error",
			bridge:  "",
			wantErr: true,
		},
		{
			name:    "bridge name with space returns error",
			bridge:  "br int",
			wantErr: true,
		},
		{
			name:    "bridge name with slash returns error",
			bridge:  "br/int",
			wantErr: true,
		},
		{
			name:    "bridge name with semicolon returns error",
			bridge:  "br;int",
			wantErr: true,
		},
		{
			name:    "bridge name with pipe returns error",
			bridge:  "br|int",
			wantErr: true,
		},
		{
			name:    "bridge name with dollar returns error",
			bridge:  "br$int",
			wantErr: true,
		},
		{
			name:    "bridge name with backtick returns error",
			bridge:  "br`int",
			wantErr: true,
		},
		{
			name:    "bridge name with special chars returns error",
			bridge:  "br@int",
			wantErr: true,
		},
		{
			name:    "bridge name with parentheses returns error",
			bridge:  "br(int)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBridgeName(tt.bridge)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBridgeName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFlowSpec(t *testing.T) {
	tests := []struct {
		name    string
		flow    string
		wantErr bool
	}{
		{
			name:    "valid flow with in_port",
			flow:    "in_port=1,icmp",
			wantErr: false,
		},
		{
			name:    "valid flow with IP addresses",
			flow:    "in_port=2,ip,nw_src=192.168.1.10,nw_dst=192.168.1.20",
			wantErr: false,
		},
		{
			name:    "valid flow with TCP ports",
			flow:    "in_port=3,tcp,nw_src=10.0.0.1,nw_dst=10.0.0.2,tp_src=12345,tp_dst=80",
			wantErr: false,
		},
		{
			name:    "valid flow with brackets",
			flow:    "in_port=1,ip,nw_src=10.244.0.5,nw_dst=10.96.0.1",
			wantErr: false,
		},
		{
			name:    "valid flow with parentheses",
			flow:    "flow(test)",
			wantErr: false,
		},
		{
			name:    "valid flow with forward slash",
			flow:    "in_port=1,ip,nw_src=10.0.0.0/24",
			wantErr: false,
		},
		{
			name:    "empty flow returns error",
			flow:    "",
			wantErr: true,
		},
		{
			name:    "flow with semicolon returns error",
			flow:    "in_port=1;icmp",
			wantErr: true,
		},
		{
			name:    "flow with ampersand returns error",
			flow:    "in_port=1&icmp",
			wantErr: true,
		},
		{
			name:    "flow with pipe returns error",
			flow:    "in_port=1|icmp",
			wantErr: true,
		},
		{
			name:    "flow with dollar returns error",
			flow:    "in_port=$1,icmp",
			wantErr: true,
		},
		{
			name:    "flow with backtick returns error",
			flow:    "in_port=`1`,icmp",
			wantErr: true,
		},
		{
			name:    "flow with less than returns error",
			flow:    "in_port=1,icmp<test",
			wantErr: true,
		},
		{
			name:    "flow with greater than returns error",
			flow:    "in_port=1,icmp>test",
			wantErr: true,
		},
		{
			name:    "flow with backslash returns error",
			flow:    "in_port=1\\icmp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFlowSpec(tt.flow)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFlowSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConntrackParams(t *testing.T) {
	tests := []struct {
		name    string
		params  []string
		wantErr bool
	}{
		{
			name:    "empty params array is valid",
			params:  []string{},
			wantErr: false,
		},
		{
			name:    "valid zone parameter",
			params:  []string{"zone=5"},
			wantErr: false,
		},
		{
			name:    "valid mark parameter with hex",
			params:  []string{"mark=0x1"},
			wantErr: false,
		},
		{
			name:    "valid labels parameter with hex",
			params:  []string{"labels=0xabcd1234"},
			wantErr: false,
		},
		{
			name:    "valid flag parameter",
			params:  []string{"-m"},
			wantErr: false,
		},
		{
			name:    "valid multiple flags",
			params:  []string{"-m", "-s"},
			wantErr: false,
		},
		{
			name:    "valid src parameter with IP",
			params:  []string{"src=10.244.0.5"},
			wantErr: false,
		},
		{
			name:    "valid dst parameter with IP",
			params:  []string{"dst=192.168.1.1"},
			wantErr: false,
		},
		{
			name:    "valid multiple parameters",
			params:  []string{"zone=5", "mark=0x1", "-m"},
			wantErr: false,
		},
		{
			name:    "valid parameter with colon",
			params:  []string{"src=10.0.0.1:8080"},
			wantErr: false,
		},
		{
			name:    "valid parameter with comma",
			params:  []string{"zone=1,2"},
			wantErr: false,
		},
		{
			name:    "valid parameter with forward slash",
			params:  []string{"src=10.0.0.0/24"},
			wantErr: false,
		},
		{
			name:    "valid parameter with underscore",
			params:  []string{"ct_zone=5"},
			wantErr: false,
		},
		{
			name:    "valid parameter with hyphen in name",
			params:  []string{"ct-zone=5"},
			wantErr: false,
		},
		{
			name:    "empty string parameter returns error",
			params:  []string{""},
			wantErr: true,
		},
		{
			name:    "parameter with semicolon returns error",
			params:  []string{"zone=5;ls"},
			wantErr: true,
		},
		{
			name:    "parameter with ampersand returns error",
			params:  []string{"zone=5&ls"},
			wantErr: true,
		},
		{
			name:    "parameter with pipe returns error",
			params:  []string{"zone=5|ls"},
			wantErr: true,
		},
		{
			name:    "parameter with dollar returns error",
			params:  []string{"zone=$USER"},
			wantErr: true,
		},
		{
			name:    "parameter with backtick returns error",
			params:  []string{"zone=`id`"},
			wantErr: true,
		},
		{
			name:    "parameter with less than returns error",
			params:  []string{"zone=5<test"},
			wantErr: true,
		},
		{
			name:    "parameter with greater than returns error",
			params:  []string{"zone=5>test"},
			wantErr: true,
		},
		{
			name:    "parameter with backslash returns error",
			params:  []string{"zone=5\\test"},
			wantErr: true,
		},
		{
			name:    "parameter with opening parenthesis returns error",
			params:  []string{"zone=(5)"},
			wantErr: true,
		},
		{
			name:    "parameter with closing parenthesis returns error",
			params:  []string{"zone=5)"},
			wantErr: true,
		},
		{
			name:    "invalid format without equals or flag returns error",
			params:  []string{"zonefive"},
			wantErr: true,
		},
		{
			name:    "invalid flag format returns error",
			params:  []string{"--zone"},
			wantErr: true,
		},
		{
			name:    "parameter with space returns error",
			params:  []string{"zone=5 test"},
			wantErr: true,
		},
		{
			name:    "one valid and one invalid parameter returns error",
			params:  []string{"zone=5", "mark=0x1;ls"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConntrackParams(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConntrackParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
