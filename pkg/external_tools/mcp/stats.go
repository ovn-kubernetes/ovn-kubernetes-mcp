package external_tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/client"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
)

func (s *MCPServer) Ethtool(ctx context.Context, req *mcp.CallToolRequest, in types.EthtoolParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := client.ValidateInterface(in.Interface); err != nil {
		return nil, types.CommandResult{}, err
	}

	operation := stringWithDefault(in.Operation, "info")
	cmd := newCommand("ethtool")

	switch operation {
	case "info":
		cmd.add(in.Interface)
	case "stats":
		cmd.add("-S", in.Interface)
	case "features":
		cmd.add("-k", in.Interface)
	default:
		return nil, types.CommandResult{}, fmt.Errorf("invalid operation: %s", operation)
	}
	return s.executeAndWrapResult(ctx, in.TargetParams, cmd.build())
}

func (s *MCPServer) ConntrackList(ctx context.Context, req *mcp.CallToolRequest, in types.ConntrackListParams) (*mcp.CallToolResult, types.CommandResult, error) {
	cmd := newCommand("conntrack", "-L").
		addIfNotEmpty(in.Protocol, "-p", in.Protocol)

	if in.SourceIP != "" {
		if err := client.ValidateIP(in.SourceIP); err != nil {
			return nil, types.CommandResult{}, fmt.Errorf("invalid source IP: %w", err)
		}
		cmd.add("--orig-src", in.SourceIP)
	}

	if in.DestIP != "" {
		if err := client.ValidateIP(in.DestIP); err != nil {
			return nil, types.CommandResult{}, fmt.Errorf("invalid destination IP: %w", err)
		}
		cmd.add("--orig-dst", in.DestIP)
	}

	if in.SourcePort > 0 {
		if err := client.ValidatePort(in.SourcePort); err != nil {
			return nil, types.CommandResult{}, err
		}
		cmd.add("--orig-port-src", strconv.Itoa(in.SourcePort))
	}

	if in.DestPort > 0 {
		if err := client.ValidatePort(in.DestPort); err != nil {
			return nil, types.CommandResult{}, err
		}
		cmd.add("--orig-port-dst", strconv.Itoa(in.DestPort))
	}

	return s.executeAndWrapResult(ctx, in.TargetParams, cmd.build())
}

func (s *MCPServer) ConntrackStats(ctx context.Context, req *mcp.CallToolRequest, in types.ConntrackStatsParams) (*mcp.CallToolResult, types.CommandResult, error) {
	command := newCommand("conntrack", "-S").build()
	return s.executeAndWrapResult(ctx, in.TargetParams, command)
}

func (s *MCPServer) SysctlNet(ctx context.Context, req *mcp.CallToolRequest, in types.SysctlNetParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := client.ValidateSysctlPattern(in.Pattern); err != nil {
		return nil, types.CommandResult{}, err
	}

	pattern := stringWithDefault(in.Pattern, "net")
	command := newCommand("sysctl", "-a", pattern).build()
	return s.executeAndWrapResult(ctx, in.TargetParams, command)
}

// registerStatsTools registers network statistics tools
func (s *MCPServer) registerStatsTools(server *mcp.Server) {
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ethtool",
			Description: `Show network interface information and statistics.

Examples:
- Show interface info: {"interface": "eth0"}
- Show statistics: {"interface": "eth0", "operation": "stats"}
- Show features: {"interface": "eth0", "operation": "features"}`,
		}, s.Ethtool)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "conntrack-list",
			Description: `List connection tracking table entries.

Examples:
- List all connections: {}
- Filter by protocol: {"protocol": "tcp"}
- Filter by source IP: {"source_ip": "192.168.1.100"}
- Filter by destination: {"dest_ip": "8.8.8.8", "dest_port": 53}`,
		}, s.ConntrackList)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "conntrack-stats",
			Description: `Show connection tracking statistics.

Example:
- Show stats: {}`,
		}, s.ConntrackStats)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "sysctl-net",
			Description: `Show network-related kernel parameters.

Examples:
- Show all network parameters: {}
- Show specific parameter: {"pattern": "net.ipv4.ip_forward"}
- Show IPv4 parameters: {"pattern": "net.ipv4"}`,
		}, s.SysctlNet)
}
