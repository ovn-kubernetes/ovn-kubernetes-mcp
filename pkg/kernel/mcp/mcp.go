// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/timeout"
)

type RunDebugNodeCommandFuncType func(ctx context.Context, namespace string, nodeName string, image string, command []string, hostPath string, mountPath string, timeout time.Duration) (string, string, error)

// Config contains the configuration for the kernel MCP server.
type Config struct {
	// Image is the container image to use for running the commands on the node.
	Image string
}

// MCPServer provides MCP server functionality for kernel operations.
type MCPServer struct {
	runDebugNodeCommand RunDebugNodeCommandFuncType
	cfg                 Config
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(runDebugNodeCommand RunDebugNodeCommandFuncType, cfg Config) (*MCPServer, error) {
	if runDebugNodeCommand == nil {
		return nil, fmt.Errorf("function to run debug node command is nil")
	}
	return &MCPServer{
		runDebugNodeCommand: runDebugNodeCommand,
		cfg:                 cfg,
	}, nil
}

// AddTools registers all kernel-related MCP tools
func (s *MCPServer) AddTools(server *mcp.Server) {
	// get-conntrack tool registration
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "get-conntrack",
			Description: fmt.Sprintf(`get-conntrack allows to interact with the connection tracking system of a Kubernetes node.
			              Use this command to discover a list of all (or a filtered selection of) currently tracked connections.
Parameters:
- node (required): Name of the node from where conntrack entries are expected to be extracted
- namespace (optional): Namespace of the debug pod from where conntrack entries are expected to be extracted. Default: 'default'
- command (optional): These options specify the particular operation to perform. These options can only be used if configured image has 'conntrack' utility available. If omitted or empty, defaults to -L. 
					  -L, --dump : List connection tracking table.
					  -C, --count: Show the table counter.
					  -S, --stats: Show the in-kernel connection tracking system statistics.
- filter_parameters (optional): Only the following read-only dump filters and output options are accepted:
                        -s, --src, --orig-src IP_ADDRESS : Match only entries whose source address in the original direction equals to mentioned IP.
						-d, --dst, --orig-dst IP_ADDRESS : Match only entries whose destination address in the original direction equals to mentioned IP.
						-p, --proto PROTO                : Specify layer four (TCP, UDP, ...) protocol.
						--sport, --orig-port-src PORT    : Source port in original direction.
						--dport, --orig-port-dst PORT    : Destination port in original direction.
Additional filters (each requires a value):
  -r/--reply-src, -q/--reply-dst, --mask-src, --mask-dst, --protonum, -f/--family,
  --reply-port-src, --reply-port-dst, --state, -m/--mark, -c/--secmark, -u/--status,
  -l/--label, -w/--zone, --orig-zone, --reply-zone, -o/--output,
  --icmp-type, --icmp-code, --icmp-id, --icmpv6-type, --icmpv6-code, --icmpv6-id,
  --srckey/--orig-key-src, --dstkey/--orig-key-dst, --reply-key-src, --reply-key-dst.
NAT filters -n/--src-nat, -g/--dst-nat, and -j/--any-nat accept an optional separate ADDRESS[:PORT] value.
Values may be separate (-p tcp), attached to short options (-ptcp), or use --name=value.
An optional full table name (conntrack, expect, dying, unconfirmed) may precede the options.
Counter clearing, alternate operations, and unlisted options or abbreviations are rejected.
Filters apply only to dump/list. Count, stats, and the /proc file fallback do not apply them.
- head (optional): Return only first N lines. Default: %d lines if tail is not specified
- tail (optional): Return only last N lines
- apply_tail_first (optional): If both head and tail are set and apply_tail_first is true,
apply tail before head. Default: false
- timeout_seconds (optional): Timeout in seconds for the command execution. If not specified, server default timeout is used. The maximum value is %d seconds.

Example:
- node='ovn-control-plane', command='-L'
- node='ovn-worker', namespace='ovn-kubernetes', command='-L'
- node='ovn-worker', filter_parameters='-s 1.2.3.4 -d 5.6.7.8 -p tcp --sport 32000 --dport 10250'

Example output:
tcp 6 91 ESTABLISHED src=1.2.3.4 dst=5.6.7.8 sport=32000 dport=10250 src=5.6.7.8 dst=1.2.3.4  sport=10250 dport=32000 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 use=2
`, DefaultMaxOutputLines, int(timeout.MaxTimeout.Seconds())),
		}, s.GetConntrack)
	// get-iptables tool registration
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "get-iptables",
			Description: fmt.Sprintf(`get-iptables allows to interact with kernel to list packet filter rules.
			              Iptables and ip6tables are used to inspect the tables of IPv4 and IPv6 packet filter rules in the Linux kernel.
Parameters:
- node (required): Name of the node from where packet filter rules are expected to be extracted
- namespace (optional): Namespace of the debug pod from where packet filter rules are expected to be extracted. Default: 'default'
- table (optional): There are currently five independent tables (which tables are present at any time depends on the kernel configuration options and which modules are present).
                    filter	: This is the default table
					nat   	: This  table is consulted when a packet that creates a new connection is encountered.
					mangle	: This table is used for specialized packet alteration.
					raw   	: This table is used mainly for configuring exemptions from connection tracking in combination with the NOTRACK target.
					security: This table is used for Mandatory Access Control (MAC) networking rules.
- command (optional): These options specify the desired action to perform. Only one of them can be specified on the command line unless otherwise stated below. If omitted or empty, defaults to -L.
					  -L, --list [chain]       : List all rules in the selected chain. If no chain is selected, all chains are listed.
					  -S, --list-rules [chain] : Print all rules in the selected chain. If no chain is selected, all chains are printed like iptables-save.
- filter_parameters (optional): An optional chain name and rule number may precede the following listing options:
								-v, --verbose					 : Verbose output.
								-n, --numeric                    : Numeric  output.   IP  addresses  and port numbers will be printed in numeric format.
								-x, --exact                      : Exact packet and byte counters.
								-w, --wait [seconds]             : Wait for the xtables lock, optionally for a non-negative number of seconds.
								--line-numbers                   : Include rule numbers.
								-4, --ipv4                       : IPv4
								-6, --ipv6                       : IPv6
Short flags may be combined (e.g., -nv4). Wait values may be separate (-w 5), attached (-w5), or use --wait=5.
Other options and abbreviated long option names are rejected. Source, destination, and protocol selectors are not valid with iptables listing commands.
Filter options apply to -L/--list. For -S/--list-rules, only the IPv4/IPv6 selection is used.
- head (optional): Return only first N lines. Default: %d lines if tail is not specified
- tail (optional): Return only last N lines
- apply_tail_first (optional): If both head and tail are set and apply_tail_first is true,
apply tail before head. Default: false
- timeout_seconds (optional): Timeout in seconds for the command execution. If not specified, server default timeout is used. The maximum value is %d seconds.
							
Example:
- node='ovn-control-plane', table='nat', filter_parameters='-nv4'
- node='ovn-control-plane', table='nat', command='-L'
Example output:
Chain POSTROUTING (policy ACCEPT 675K packets, 41M bytes)
 pkts bytes target     prot opt in     out     source               destination         
 675K   41M OVN-KUBE-EGRESS-IP-MULTI-NIC  all  --  *      *       0.0.0.0/0            0.0.0.0/0     							
`, DefaultMaxOutputLines, int(timeout.MaxTimeout.Seconds())),
		}, s.GetIptables)
	// get-nft tool registration
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "get-nft",
			Description: fmt.Sprintf(`get-nft allows to interact with kernel to list packet filtering and classification rules.
Parameters:
- node (required): Name of the node from where packet filtering and classification rules are expected to be extracted
- namespace (optional): Namespace of the debug pod from where packet filtering and classification rules are expected to be extracted. Default: 'default'
- command (required): These options specify the desired action to perform. Only one of them can be specified on the command line unless otherwise stated below.
                    - list ruleset   : The ruleset keyword is used to identify the whole set of tables, chains, etc. Print the ruleset in human-readable format.
					- list tables    : List all chains and rules of the specified table.
					- list chains    : List all rules of the specified chain.
					- list sets      : Display the elements in the specified set.
					- list maps      : Display the elements in the specified map.
					- list flowtables: List all flowtables.
- address_families (optional): Address families determine the type of packets which are processed. For each address family, the kernel contains so-called hooks at specific stages of
       						   the packet processing paths, which invoke nftables if rules for these hooks exist.
							   - ip       IPv4 address family.
                               - ip6      IPv6 address family.
                               - inet     Internet (IPv4/IPv6) address family.
                               - arp      ARP address family, handling IPv4 ARP packets.
                               - bridge   Bridge address family, handling packets which traverse a bridge device.
                               - netdev   Netdev address family, handling packets on ingress and egress.
- head (optional): Return only first N lines. Default: %d lines if tail is not specified
- tail (optional): Return only last N lines
- apply_tail_first (optional): If both head and tail are set and apply_tail_first is true,
apply tail before head. Default: false
- timeout_seconds (optional): Timeout in seconds for the command execution. If not specified, server default timeout is used. The maximum value is %d seconds.
					
Example:
- node='ovn-control-plane', command='list tables', address_families='inet'
Example output:
table inet ovn-kubernetes
`, DefaultMaxOutputLines, int(timeout.MaxTimeout.Seconds())),
		}, s.GetNFT)
	// get-ip tool registration
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "get-ip",
			Description: fmt.Sprintf(`get-ip allows to interact with kernel to list routing, network devices, interfaces.
Parameters:
- node (required): Name of the node on which ip command is expected to be executed
- namespace (optional): Namespace of the debug pod on which ip command is expected to be executed. Default: 'default'
- options (optional): Diagnostic global options, using the following exact names with one or two leading hyphens:
                      -d, -details; -s, -stats, -statistics; -o, -oneline; -r, -resolve; -a, -all; -N, -Numeric.
                      -j, -json; -p, -pretty; -br, -brief; -h, -human, -human-readable; -iec.
                      -t, -timestamp; -ts, -tshort; -c, -color[=always|auto|never].
                      -4, -6, -0, -B, -M; -f, -family <inet|inet6|link|bridge|mpls>.
                      -n, -netns <NETNS>; -rc, -rcvbuf <SIZE> (unsigned 32-bit integer).
Namespace, family, and buffer options require a separate value. Short flags must be separate (e.g., -s -d, not -sd).
Batch mode, other options, and unlisted abbreviations are rejected.
- command (required): These options specify the desired action to perform. Only one of them can be specified on the command line unless otherwise stated below.
                      - address show     : protocol (IP or IPv6) address on a device.
					  - link show        : network device.
					  - neighbour show   : manage ARP or NDISC cache entries.
					  - netns show       : manage network namespaces.
					  - route show       : routing table entry.
					  - rule show        : rule in routing policy database.
					  - vrf show         : manage virtual routing and forwarding devices. 
					  - xfrm state list  : show Security Association Database.
					  - xfrm policy list : show Security Policy Database.
- filter_parameters (optional): This allows to mention sub command to get more filtered data. Available sub command varies and supportability depends on what is 
                          already supported with 'ip' utility.
- head (optional): Return only first N lines. Default: %d lines if tail is not specified
- tail (optional): Return only last N lines
- apply_tail_first (optional): If both head and tail are set and apply_tail_first is true,
apply tail before head. Default: false
- timeout_seconds (optional): Timeout in seconds for the command execution. If not specified, server default timeout is used. The maximum value is %d seconds.

Example:
- node='ovn-control-plane', options="-4", command='route show', filter_parameters='table all'
Example output:
default via 10.0.0.254 dev br-ex proto dhcp src 10.0.0.10 metric 48
`, DefaultMaxOutputLines, int(timeout.MaxTimeout.Seconds())),
		}, s.GetIPCommandOutput)
}

// executeCommand executes a command on a node via kubectl debug
func (s *MCPServer) executeCommand(ctx context.Context, namespace, node string, command []string) (string, string, error) {
	stdout, stderr, err := s.runDebugNodeCommand(ctx, namespace, node, s.cfg.Image, command, "", "", 0)
	if err != nil {
		return "", "", fmt.Errorf("error while establishing tty connection to the node: %w", err)
	}

	// Filter out warning lines from stderr
	if stderr != "" {
		stderr = filterWarnings(stderr)
	}

	// Filter out warning lines from stdout
	stdout = filterWarnings(stdout)

	return stdout, stderr, nil
}
