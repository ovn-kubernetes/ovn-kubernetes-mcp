// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kernel/types"
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

func TestGetIPRejectsUnsafeOptions(t *testing.T) {
	for _, options := range []string{
		"-batch /host/tmp/commands", "--batch /host/tmp/commands", "-b /host/tmp/commands",
		"-ba /host/tmp/commands", "--bat /host/tmp/commands", "-batch -", "--batch=/host/tmp/commands",
		"-br -batch /host/tmp/commands", "-force", "link set lo down", "-- link set lo down",
		"-", "--", "-de", "-sta", "-j -b /host/tmp/commands", "-sd", "-jp", "-ss",
		"-n", "--netns", "-netns -batch /host/tmp/commands", "-n blue link set lo down",
		"-n blue\x00", "-netns=blue", "-f", "--family", "-family -batch /host/tmp/commands",
		"-family unknown", "-family help", "-family inet6 link set lo down", "-f=inet",
		"-rcvbuf", "-rc -1", "-rcvbuf abc", "-rcvbuf 4294967296", "-rcvbuf -batch /host/tmp/commands",
		"-color=blue", "-color never", "-json=true", "-details; true", "-unknown",
	} {
		t.Run(options, func(t *testing.T) {
			var commands [][]string
			s := &MCPServer{runDebugNodeCommand: func(_ context.Context, _, _, _ string, cmd []string, _, _ string, _ time.Duration) (string, string, error) {
				commands = append(commands, cmd)
				return "output", "", nil
			}}
			_, _, err := s.GetIPCommandOutput(t.Context(), nil, types.ListIPParams{Command: "link show", Options: options})
			if err == nil {
				t.Errorf("unsafe options accepted; executed %q", commands)
			}
			if len(commands) != 0 {
				t.Errorf("invalid options reached debug pod execution: %q", commands)
			}
		})
	}
}

func TestGetIPCommand(t *testing.T) {
	tests := []struct {
		name    string
		options string
		command string
		filters string
		want    []string
	}{
		{"defaults", "", "route show", "", []string{"ip", "route", "show"}},
		{"documented route", "-4", "route show", "table all", []string{"ip", "-4", "route", "show", "table", "all"}},
		{"short display options", "-s -s -d -o -N -r -a", "link show", "dev lo", []string{"ip", "-s", "-s", "-d", "-o", "-N", "-r", "-a", "link", "show", "dev", "lo"}},
		{"long display options", "-stats -statistics -details -oneline -Numeric -resolve -all", "address show", "", []string{"ip", "-stats", "-statistics", "-details", "-oneline", "-Numeric", "-resolve", "-all", "address", "show"}},
		{"JSON formatting", "-j -p", "link show", "", []string{"ip", "-j", "-p", "link", "show"}},
		{"long JSON formatting", "--json --pretty --details", "link show", "", []string{"ip", "--json", "--pretty", "--details", "link", "show"}},
		{"brief output", "-br -6", "address show", "", []string{"ip", "-br", "-6", "address", "show"}},
		{"long brief output", "-brief", "link show", "", []string{"ip", "-brief", "link", "show"}},
		{"namespace", "-n tenant-blue -d", "route show", "", []string{"ip", "-n", "tenant-blue", "-d", "route", "show"}},
		{"long namespace", "--netns tenant_blue", "rule show", "", []string{"ip", "--netns", "tenant_blue", "rule", "show"}},
		{"family", "-f inet", "route show", "", []string{"ip", "-f", "inet", "route", "show"}},
		{"long family", "--family inet6", "route show", "", []string{"ip", "--family", "inet6", "route", "show"}},
		{"link family", "-family link", "neighbour show", "", []string{"ip", "-family", "link", "neighbour", "show"}},
		{"bridge family", "-family bridge", "link show", "", []string{"ip", "-family", "bridge", "link", "show"}},
		{"MPLS family", "-family mpls", "route show", "", []string{"ip", "-family", "mpls", "route", "show"}},
		{"family shortcuts", "-0 -B -M", "route show", "", []string{"ip", "-0", "-B", "-M", "route", "show"}},
		{"human readable", "-h -human -human-readable -iec", "link show", "", []string{"ip", "-h", "-human", "-human-readable", "-iec", "link", "show"}},
		{"timestamps", "-t -timestamp -ts -tshort", "link show", "", []string{"ip", "-t", "-timestamp", "-ts", "-tshort", "link", "show"}},
		{"colors", "-c -color=auto -c=never --color=always -c=", "link show", "", []string{"ip", "-c", "-color=auto", "-c=never", "--color=always", "-c=", "link", "show"}},
		{"receive buffer", "-rc 1048576 -rcvbuf 0x100000", "route show", "", []string{"ip", "-rc", "1048576", "-rcvbuf", "0x100000", "route", "show"}},
		{"whitespace and command abbreviation", "\t-d  -o ", "addr sh", "dev lo", []string{"ip", "-d", "-o", "addr", "sh", "dev", "lo"}},
		{"xfrm command", "-s", "xfrm state list", "", []string{"ip", "-s", "xfrm", "state", "list"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var commands [][]string
			s := &MCPServer{runDebugNodeCommand: func(_ context.Context, _, _, _ string, cmd []string, _, _ string, _ time.Duration) (string, string, error) {
				commands = append(commands, cmd)
				return "output", "", nil
			}}
			_, _, err := s.GetIPCommandOutput(t.Context(), nil, types.ListIPParams{
				Options: tt.options, Command: tt.command, FilterParameters: tt.filters,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(commands) != 2 {
				t.Fatalf("commands = %q, want availability probe and display command", commands)
			}
			if want := []string{"ip", "-V"}; !slices.Equal(commands[0], want) {
				t.Errorf("availability probe = %q, want %q", commands[0], want)
			}
			if !slices.Equal(commands[1], tt.want) {
				t.Errorf("command = %q, want %q", commands[1], tt.want)
			}
		})
	}
}

func TestGetIPRejectsMetacharacters(t *testing.T) {
	s := &MCPServer{runDebugNodeCommand: func(_ context.Context, _, _, _ string, _ []string, _, _ string, _ time.Duration) (string, string, error) {
		t.Fatal("invalid options reached debug pod execution")
		return "", "", nil
	}}
	_, _, err := s.GetIPCommandOutput(t.Context(), nil, types.ListIPParams{Command: "link show", Options: "-d; true"})
	if err == nil || !strings.Contains(err.Error(), "invalid use of metacharacters in parameter") {
		t.Errorf("error = %v, want existing metacharacter validation error", err)
	}
}
