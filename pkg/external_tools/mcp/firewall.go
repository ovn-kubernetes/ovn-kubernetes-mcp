package external_tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/client"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
)

func (s *MCPServer) IPTablesList(ctx context.Context, req *mcp.CallToolRequest, in types.IPTablesListParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := client.ValidateTableName(in.Table); err != nil {
		return nil, types.CommandResult{}, err
	}

	cmdName := "iptables"
	if in.IPv6 {
		cmdName = "ip6tables"
	}

	table := stringWithDefault(in.Table, "filter")

	cmd := newCommand(cmdName, "-L").
		addIfNotEmpty(in.Chain, in.Chain).
		add("-t", table).
		addIf(in.LineNumbers, "--line-numbers").
		addIf(in.Verbose, "-v").
		add("-n")
	return s.executeAndWrapResult(ctx, in.TargetParams, cmd.build())
}

func (s *MCPServer) NFTList(ctx context.Context, req *mcp.CallToolRequest, in types.NFTListParams) (*mcp.CallToolResult, types.CommandResult, error) {
	cmd := newCommand("nft", "list")

	if in.Table != "" && in.Chain != "" {
		cmd.add("chain", in.Table, in.Chain)
	} else if in.Table != "" {
		cmd.add("table", in.Table)
	} else {
		cmd.add("ruleset")
	}
	return s.executeAndWrapResult(ctx, in.TargetParams, cmd.build())
}

// registerFirewallTools registers firewall inspection tools
func (s *MCPServer) registerFirewallTools(server *mcp.Server) {
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "iptables-list",
			Description: `List iptables/ip6tables firewall rules (read-only).

Examples:
- List all filter rules: {}
- List NAT table: {"table": "nat"}
- List with line numbers: {"line_numbers": true}
- List IPv6 rules: {"ipv6": true}
- List specific chain: {"chain": "INPUT", "verbose": true}`,
		}, s.IPTablesList)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "nft-list",
			Description: `List nftables rules (read-only).

Examples:
- List all rules: {}
- List specific table: {"table": "inet filter"}
- List specific chain: {"table": "inet filter", "chain": "input"}`,
		}, s.NFTList)
}
