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

func TestGetIptablesRejectsUnsafeFilters(t *testing.T) {
	filters := []string{
		"-Z", "--zero", "-nvZ", "-A INPUT", "-F", "-t security", "--table=security",
		"--modprobe=/host/usr/bin/dnf", "--modprobe /host/usr/bin/dnf", "-M/host/usr/bin/dnf",
		"--modp=/host/usr/bin/dnf", "--numeric=true", "--line-numbers=true", "--", "-",
		"-s", "-s -n", "--source=", "-p --modprobe=/host/usr/bin/dnf", "-p--modprobe=/host/usr/bin/dnf",
		"--source=-Z", "--protocol=-Z", "-w-1", "-wINPUT", "-w6Z", "--wait=-1", "--wait=abc", "--wait=", "-w 4294967296",
		"-p /host/protocol", "-s /host/address", "INPUT OUTPUT", "-m state", "--dport 80",
		"-n; true", "INPUT\x00", "--unknown", "-n6Z", "-n0",
		"-p6", "-nvp6", "-s 10.0.0.1", "--src=10.0.0.1", "--dst=192.0.2.1",
		"INPUT 0", "INPUT -1", "INPUT 2147483648", "INPUT 3 4", "-w INPUT",
	}
	for _, command := range []string{"", "-S"} {
		for _, filter := range filters {
			t.Run(command+"/"+filter, func(t *testing.T) {
				s := &MCPServer{runDebugNodeCommand: func(_ context.Context, _, _, _ string, _ []string, _, _ string, _ time.Duration) (string, string, error) {
					t.Fatal("invalid filters reached debug pod execution")
					return "", "", nil
				}}
				_, _, err := s.GetIptables(t.Context(), nil, types.ListIPTablesParams{
					Command: command, FilterParameters: filter,
				})
				if err == nil {
					t.Fatal("expected invalid filter parameters to be rejected")
				}
			})
		}
	}
}

func TestGetIptablesRejectsIncompatibleFilters(t *testing.T) {
	for _, command := range []string{"-S", "--list-rules", " --list-rules "} {
		for _, filters := range []string{"-n", "--numeric", "-x", "--exact", "--line-numbers", "INPUT -nv6"} {
			t.Run(command+"/"+filters, func(t *testing.T) {
				s := &MCPServer{runDebugNodeCommand: func(_ context.Context, _, _, _ string, _ []string, _, _ string, _ time.Duration) (string, string, error) {
					t.Fatal("incompatible filters reached debug pod execution")
					return "", "", nil
				}}
				_, _, err := s.GetIptables(t.Context(), nil, types.ListIPTablesParams{
					Command: command, FilterParameters: filters,
				})
				if err == nil || !strings.Contains(err.Error(), "is not supported with "+strings.TrimSpace(command)) {
					t.Fatalf("error = %v, want incompatible filter error", err)
				}
			})
		}
	}
}

func TestGetIptablesCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		table   string
		filters string
		want    []string
	}{
		{"defaults", "", "", "", []string{"iptables", "-t", "filter", "-L"}},
		{"documented bundle", "", "nat", "-nv4", []string{"iptables", "-t", "nat", "-L", "-nv4"}},
		{"IPv6 bundle", "--list", "", "-nv6", []string{"ip6tables", "-t", "filter", "--list", "-nv6"}},
		{"IPv6 long flag", "", "", "--ipv6 --numeric", []string{"ip6tables", "-t", "filter", "-L", "--ipv6", "--numeric"}},
		{"IPv6 tab separator", "", "", "-n\t-6", []string{"ip6tables", "-t", "filter", "-L", "-n", "-6"}},
		{"chain with hyphen and six", "", "", "CUSTOM-6", []string{"iptables", "-t", "filter", "-L", "CUSTOM-6"}},
		{"wait value", "", "", "-w6", []string{"iptables", "-t", "filter", "-L", "-w6"}},
		{"wait long value", "", "", "--wait=6", []string{"iptables", "-t", "filter", "-L", "--wait=6"}},
		{"wait separate value", "", "", "INPUT -w 6", []string{"iptables", "-t", "filter", "-L", "INPUT", "-w", "6"}},
		{"wait without value", "", "", "-w", []string{"iptables", "-t", "filter", "-L", "-w"}},
		{"listing controls", "", "", "INPUT -vnx --line-numbers", []string{"iptables", "-t", "filter", "-L", "INPUT", "-vnx", "--line-numbers"}},
		{"long listing controls", "", "", "--verbose --numeric --exact", []string{"iptables", "-t", "filter", "-L", "--verbose", "--numeric", "--exact"}},
		// iptables(8) and upstream xshared.c allow -L [chain [rulenum]].
		{"chain and rule", "", "", "INPUT 3 -nv4", []string{"iptables", "-t", "filter", "-L", "INPUT", "3", "-nv4"}},
		{"IPv6 chain and rule", "", "", "INPUT 3 -nv6", []string{"ip6tables", "-t", "filter", "-L", "INPUT", "3", "-nv6"}},
		{"chain punctuation", "", "", "tenant@blue+1 -n", []string{"iptables", "-t", "filter", "-L", "tenant@blue+1", "-n"}},
		{"repeated verbose", "", "", "-vvn4", []string{"iptables", "-t", "filter", "-L", "-vvn4"}},
		{"native numeric wait", "", "", "-w0x6", []string{"iptables", "-t", "filter", "-L", "-w0x6"}},
		{"rule listing", "--list-rules", "raw", "", []string{"iptables", "-t", "raw", "--list-rules"}},
		{"rule listing chain", "-S", "", "INPUT", []string{"iptables", "-t", "filter", "-S", "INPUT"}},
		{"long rule listing chain", "--list-rules", "raw", "PREROUTING --wait=1", []string{"iptables", "-t", "raw", "--list-rules", "PREROUTING", "--wait=1"}},
		{"IPv6 rule listing chain", "-S", "", "INPUT -6", []string{"ip6tables", "-t", "filter", "-S", "INPUT", "-6"}},
		{"rule listing options", "-S", "", "-6 -v", []string{"ip6tables", "-t", "filter", "-S", "-6", "-v"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var commands [][]string
			s := &MCPServer{runDebugNodeCommand: func(_ context.Context, _, _, _ string, cmd []string, _, _ string, _ time.Duration) (string, string, error) {
				commands = append(commands, cmd)
				return "Chain INPUT", "", nil
			}}
			_, _, err := s.GetIptables(t.Context(), nil, types.ListIPTablesParams{
				Command: tt.command, Table: tt.table, FilterParameters: tt.filters,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(commands) != 2 {
				t.Fatalf("commands = %q, want availability probe and listing", commands)
			}
			if want := []string{tt.want[0], "-V"}; !slices.Equal(commands[0], want) {
				t.Errorf("availability probe = %q, want %q", commands[0], want)
			}
			if !slices.Equal(commands[1], tt.want) {
				t.Errorf("command = %q, want %q", commands[1], tt.want)
			}
		})
	}
}

func TestGetIptablesRejectsMetacharacters(t *testing.T) {
	s := &MCPServer{runDebugNodeCommand: func(_ context.Context, _, _, _ string, _ []string, _, _ string, _ time.Duration) (string, string, error) {
		t.Fatal("invalid filters reached debug pod execution")
		return "", "", nil
	}}
	_, _, err := s.GetIptables(t.Context(), nil, types.ListIPTablesParams{FilterParameters: "-n; true"})
	if err == nil || !strings.Contains(err.Error(), "invalid use of metacharacters in parameter") {
		t.Errorf("error = %v, want existing metacharacter validation error", err)
	}
}
