// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kernel/types"
)

func TestValidateConntrackCommand(t *testing.T) {
	tests := []struct {
		name         string
		command      string
		cliAvailable bool
		wantError    bool
	}{
		{
			name:         "empty string with CLI available",
			command:      "",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "empty string without CLI",
			command:      "",
			cliAvailable: false,
			wantError:    false,
		},
		{
			name:         "numeric string with CLI available",
			command:      "321",
			cliAvailable: true,
			wantError:    true,
		},
		{
			name:         "numeric string without CLI",
			command:      "321",
			cliAvailable: false,
			wantError:    true,
		},
		{
			name:         "valid command --dump with CLI available",
			command:      "--dump",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "valid command --dump without CLI",
			command:      "--dump",
			cliAvailable: false,
			wantError:    false,
		},
		{
			name:         "valid command --stats with CLI available",
			command:      "--stats",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "valid command --stats without CLI",
			command:      "--stats",
			cliAvailable: false,
			wantError:    true,
		},
		{
			name:         "valid command --count with CLI available",
			command:      "--count",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "valid command --count without CLI",
			command:      "--count",
			cliAvailable: false,
			wantError:    true,
		},
		{
			name:         "valid command -L with CLI available",
			command:      "-L",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "valid command -L without CLI",
			command:      "-L",
			cliAvailable: false,
			wantError:    false,
		},
		{
			name:         "valid command -S with CLI available",
			command:      "-S",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "valid command -S without CLI",
			command:      "-S",
			cliAvailable: false,
			wantError:    true,
		},
		{
			name:         "valid command -C with CLI available",
			command:      "-C",
			cliAvailable: true,
			wantError:    false,
		},
		{
			name:         "valid command -C without CLI",
			command:      "-C",
			cliAvailable: false,
			wantError:    true,
		},
		{
			name:         "invalid command with CLI available",
			command:      "-E",
			cliAvailable: true,
			wantError:    true,
		},
		{
			name:         "invalid command without CLI",
			command:      "-E",
			cliAvailable: false,
			wantError:    true,
		},
		{
			name:         "invalid command -D with CLI available",
			command:      "-D",
			cliAvailable: true,
			wantError:    true,
		},
		{
			name:         "lowercase -l with CLI available",
			command:      "-l",
			cliAvailable: true,
			wantError:    true,
		},
		{
			name:         "command with spaces",
			command:      "-L ",
			cliAvailable: true,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConntrackCommand(tt.command, tt.cliAvailable)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateConntrackCommand(%q, %v) error = %v, wantError %v", tt.command, tt.cliAvailable, err, tt.wantError)
			}
		})
	}
}

func TestGetConntrackRejectsUnsafeFilters(t *testing.T) {
	filters := []string{
		"-z", "--zero", "--zer", "-zn", "-zz", "-o extended -z", "conntrack --zero",
		"-n -z", "--src-nat --zero", "-n192.0.2.1 -z", "--src-nat=192.0.2.1 --zero",
		"-D", "--delete", "-I", "--create", "-A", "--add", "-U", "--update", "-F", "--flush",
		"-E", "--event", "-G", "--get", "-R /host/tmp/entries", "--load-file /host/tmp/entries",
		"-L", "--dump", "-S", "--stats", "-C", "--count", "--help", "--version",
		"--label-add test", "--label-del test", "--timeout 30", "--event-mask NEW", "--buffer-size 1024",
		"--orig-vtag 1", "--role client", "--tuple-src 192.0.2.1", "--mask-port-src 65535",
		"-s", "-s -z", "--src=", "--src=-z", "-p-z", "--sport --zero", "--output=--zero",
		"--orig-zone -1", "--src-nat=-z", "--any-nat=", "-n-z", "--prot tcp", "--unknown",
		"--", "-", "expect -z", "invalid-table", "-p tcp extra", "--state ESTABLISHED\x00", "-s 192.0.2.1; true",
	}
	for _, command := range []string{"", "-L", "--dump", "-S", "-C"} {
		for _, filter := range filters {
			t.Run(command+"/"+filter, func(t *testing.T) {
				var commands [][]string
				s := &MCPServer{runDebugNodeCommand: func(_ context.Context, _, _, _ string, cmd []string, _, _ string, _ time.Duration) (string, string, error) {
					commands = append(commands, cmd)
					return "output", "", nil
				}}
				_, _, err := s.GetConntrack(t.Context(), nil, types.ListConntrackParams{
					Command: command, FilterParameters: filter,
				})
				if err == nil {
					t.Errorf("unsafe filters accepted; executed %q", commands)
				}
				if len(commands) != 0 {
					t.Errorf("invalid filters reached debug pod execution: %q", commands)
				}
			})
		}
	}
}

func TestGetConntrackCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		filters string
		want    []string
	}{
		{"defaults", "", "", []string{"conntrack", "-L"}},
		{"documented filters", "-L", "-s 1.2.3.4 -d 5.6.7.8 -p tcp --sport 32000 --dport 10250", []string{"conntrack", "-L", "-s", "1.2.3.4", "-d", "5.6.7.8", "-p", "tcp", "--sport", "32000", "--dport", "10250"}},
		{"long filters", "--dump", "--src 192.0.2.1 --orig-dst 198.51.100.1 --proto tcp --orig-port-src 1234 --orig-port-dst 443", []string{"conntrack", "--dump", "--src", "192.0.2.1", "--orig-dst", "198.51.100.1", "--proto", "tcp", "--orig-port-src", "1234", "--orig-port-dst", "443"}},
		{"attached values", "", "-ptcp --orig-src=192.0.2.0/24 --dst=198.51.100.1 --dport=443 -oextended,id", []string{"conntrack", "-L", "-ptcp", "--orig-src=192.0.2.0/24", "--dst=198.51.100.1", "--dport=443", "-oextended,id"}},
		{"IPv6 reply filters", "", "-f ipv6 -r 2001:db8::1 -q 2001:db8::2 -p udp --reply-port-src 53 --reply-port-dst 1234", []string{"conntrack", "-L", "-f", "ipv6", "-r", "2001:db8::1", "-q", "2001:db8::2", "-p", "udp", "--reply-port-src", "53", "--reply-port-dst", "1234"}},
		{"long reply filters", "", "--family ipv4 --reply-src 192.0.2.1 --reply-dst 198.51.100.1 --protonum 6", []string{"conntrack", "-L", "--family", "ipv4", "--reply-src", "192.0.2.1", "--reply-dst", "198.51.100.1", "--protonum", "6"}},
		{"mark and status", "", "-m 0x1/0xff -c 2 -u ASSURED -w 3", []string{"conntrack", "-L", "-m", "0x1/0xff", "-c", "2", "-u", "ASSURED", "-w", "3"}},
		{"long metadata filters", "", "--mark 1 --secmark 2 --status !SEEN_REPLY --zone 3 --orig-zone 4 --reply-zone 5", []string{"conntrack", "-L", "--mark", "1", "--secmark", "2", "--status", "!SEEN_REPLY", "--zone", "3", "--orig-zone", "4", "--reply-zone", "5"}},
		{"labels and masks", "", "-l test --label blue --mask-src 255.255.255.0 --mask-dst 255.255.255.0", []string{"conntrack", "-L", "-l", "test", "--label", "blue", "--mask-src", "255.255.255.0", "--mask-dst", "255.255.255.0"}},
		{"source NAT", "", "--src-nat -o extended", []string{"conntrack", "-L", "--src-nat", "-o", "extended"}},
		{"destination NAT address", "", "-g 192.0.2.1:443 --output extended,id", []string{"conntrack", "-L", "-g", "192.0.2.1:443", "--output", "extended,id"}},
		{"any NAT", "", "--any-nat", []string{"conntrack", "-L", "--any-nat"}},
		{"NAT short flags", "", "-n -g", []string{"conntrack", "-L", "-n", "-g"}},
		{"any NAT address", "", "-j 192.0.2.1", []string{"conntrack", "-L", "-j", "192.0.2.1"}},
		{"ICMP", "", "-p icmp --icmp-type 8 --icmp-code 0 --icmp-id 123", []string{"conntrack", "-L", "-p", "icmp", "--icmp-type", "8", "--icmp-code", "0", "--icmp-id", "123"}},
		{"ICMPv6", "", "-p icmpv6 --icmpv6-type 128 --icmpv6-code 0 --icmpv6-id 123", []string{"conntrack", "-L", "-p", "icmpv6", "--icmpv6-type", "128", "--icmpv6-code", "0", "--icmpv6-id", "123"}},
		{"state", "", "-p tcp --state ESTABLISHED", []string{"conntrack", "-L", "-p", "tcp", "--state", "ESTABLISHED"}},
		{"GRE", "", "-p gre --srckey 1 --dstkey 2 --reply-key-src 3 --reply-key-dst 4", []string{"conntrack", "-L", "-p", "gre", "--srckey", "1", "--dstkey", "2", "--reply-key-src", "3", "--reply-key-dst", "4"}},
		{"long GRE keys", "", "-p gre --orig-key-src 1 --orig-key-dst 2", []string{"conntrack", "-L", "-p", "gre", "--orig-key-src", "1", "--orig-key-dst", "2"}},
		{"expectation table", "", "expect -o xml", []string{"conntrack", "-L", "expect", "-o", "xml"}},
		{"dying table", "", "dying", []string{"conntrack", "-L", "dying"}},
		{"unconfirmed table", "", "unconfirmed", []string{"conntrack", "-L", "unconfirmed"}},
		{"explicit default table", "", "conntrack -o save", []string{"conntrack", "-L", "conntrack", "-o", "save"}},
		{"statistics omit filters", "--stats", "-p tcp", []string{"conntrack", "--stats"}},
		{"count omits filters", "--count", "-p udp", []string{"conntrack", "--count"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var commands [][]string
			s := &MCPServer{runDebugNodeCommand: func(_ context.Context, _, _, _ string, cmd []string, _, _ string, _ time.Duration) (string, string, error) {
				commands = append(commands, cmd)
				return "output", "", nil
			}}
			_, _, err := s.GetConntrack(t.Context(), nil, types.ListConntrackParams{Command: tt.command, FilterParameters: tt.filters})
			if err != nil {
				t.Fatal(err)
			}
			if len(commands) != 2 {
				t.Fatalf("commands = %q, want availability probe and listing", commands)
			}
			if want := []string{"conntrack", "-V"}; !slices.Equal(commands[0], want) {
				t.Errorf("availability probe = %q, want %q", commands[0], want)
			}
			if !slices.Equal(commands[1], tt.want) {
				t.Errorf("command = %q, want %q", commands[1], tt.want)
			}
		})
	}
}

func TestGetConntrackFallback(t *testing.T) {
	for _, command := range []string{"", "-L", "--dump", "-S", "-C"} {
		t.Run(command, func(t *testing.T) {
			var commands [][]string
			s := &MCPServer{runDebugNodeCommand: func(_ context.Context, _, _, _ string, cmd []string, _, _ string, _ time.Duration) (string, string, error) {
				commands = append(commands, cmd)
				if cmd[0] == "conntrack" {
					return "", "", errors.New("conntrack unavailable")
				}
				return "tcp src=192.0.2.1", "", nil
			}}
			_, _, err := s.GetConntrack(t.Context(), nil, types.ListConntrackParams{Command: command, FilterParameters: "-p tcp"})
			if command == "-S" || command == "-C" {
				if err == nil || !strings.Contains(err.Error(), "only -L/--dump is supported") || len(commands) != 1 {
					t.Fatalf("error = %v, commands = %q; want unsupported operation before file read", err, commands)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(commands) != 2 || !slices.Equal(commands[1], []string{"cat", conntrackSystemFile}) {
				t.Errorf("commands = %q, want fallback file read without CLI filters", commands)
			}
		})
	}
}
