package external_tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/client"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
)

func (s *MCPServer) IPAddrShow(ctx context.Context, req *mcp.CallToolRequest, in types.IPAddrShowParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := client.ValidateInterface(in.Interface); err != nil {
		return nil, types.CommandResult{}, err
	}
	command := newCommand("ip", "addr", "show").
		addIfNotEmpty(in.Interface, "dev", in.Interface).
		build()
	return s.executeAndWrapResult(ctx, in.TargetParams, command)
}

func (s *MCPServer) IPRouteShow(ctx context.Context, req *mcp.CallToolRequest, in types.IPRouteShowParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := client.ValidateTableName(in.Table); err != nil {
		return nil, types.CommandResult{}, err
	}
	command := newCommand("ip", "route", "show").
		addIfNotEmpty(in.Table, "table", in.Table).
		addIfNotEmpty(in.Destination, in.Destination).
		build()
	return s.executeAndWrapResult(ctx, in.TargetParams, command)
}

func (s *MCPServer) IPLinkShow(ctx context.Context, req *mcp.CallToolRequest, in types.IPLinkShowParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := client.ValidateInterface(in.Interface); err != nil {
		return nil, types.CommandResult{}, err
	}
	command := newCommand("ip", "link", "show").
		addIfNotEmpty(in.Interface, "dev", in.Interface).
		build()
	return s.executeAndWrapResult(ctx, in.TargetParams, command)
}

func (s *MCPServer) IPNeighShow(ctx context.Context, req *mcp.CallToolRequest, in types.IPNeighShowParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := client.ValidateInterface(in.Interface); err != nil {
		return nil, types.CommandResult{}, err
	}
	if err := client.ValidateIP(in.Address); err != nil {
		return nil, types.CommandResult{}, err
	}
	command := newCommand("ip", "neigh", "show").
		addIfNotEmpty(in.Interface, "dev", in.Interface).
		addIfNotEmpty(in.Address, in.Address).
		build()
	return s.executeAndWrapResult(ctx, in.TargetParams, command)
}

func (s *MCPServer) IPRuleShow(ctx context.Context, req *mcp.CallToolRequest, in types.IPRuleShowParams) (*mcp.CallToolResult, types.CommandResult, error) {
	command := newCommand("ip", "rule", "show").build()
	return s.executeAndWrapResult(ctx, in.TargetParams, command)
}

// registerIPTools registers all IP command tools
func (s *MCPServer) registerIPTools(server *mcp.Server) {
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ip-addr-show",
			Description: `Show IP addresses on network interfaces.

Examples:
- Show all interfaces: {}
- Show specific interface: {"interface": "eth0"}`,
		}, s.IPAddrShow)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ip-route-show",
			Description: `Show the IP routing table.

Examples:
- Show all routes: {}
- Show specific destination: {"destination": "10.0.0.0/8"}
- Show specific table: {"table": "main"}`,
		}, s.IPRouteShow)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ip-link-show",
			Description: `Show network interfaces and their status.

Examples:
- Show all interfaces: {}
- Show specific interface: {"interface": "eth0"}`,
		}, s.IPLinkShow)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ip-neigh-show",
			Description: `Show ARP/neighbor cache entries.

Examples:
- Show all neighbors: {}
- Show neighbors on interface: {"interface": "eth0"}
- Show specific address: {"address": "192.168.1.1"}`,
		}, s.IPNeighShow)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ip-rule-show",
			Description: `Show routing policy database (RPDB) rules.

Example:
- Show all rules: {}`,
		}, s.IPRuleShow)
}
