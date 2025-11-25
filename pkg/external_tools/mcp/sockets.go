package external_tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
)

func (s *MCPServer) SS(ctx context.Context, req *mcp.CallToolRequest, in types.SSParams) (*mcp.CallToolResult, types.CommandResult, error) {
	cmd := newCommand("ss")

	switch in.Protocol {
	case "tcp":
		cmd.add("-t")
	case "udp":
		cmd.add("-u")
	case "all", "":
		cmd.add("-a")
	}

	switch in.State {
	case "listening":
		cmd.add("-l")
	case "established":
		cmd.add("state", "established")
	}

	cmd.addIf(in.Process, "-p").
		addIf(in.Numeric, "-n").
		addIfNotEmpty(in.PortFilter, in.PortFilter)
	return s.executeAndWrapResult(ctx, in.TargetParams, cmd.build())
}

func (s *MCPServer) Netstat(ctx context.Context, req *mcp.CallToolRequest, in types.NetstatParams) (*mcp.CallToolResult, types.CommandResult, error) {
	cmd := newCommand("netstat")

	switch in.Protocol {
	case "tcp":
		cmd.add("-t")
	case "udp":
		cmd.add("-u")
	case "all", "":
		cmd.add("-a")
	}

	cmd.addIf(in.Listening, "-l").
		addIf(in.Numeric, "-n").
		add("-p")
	return s.executeAndWrapResult(ctx, in.TargetParams, cmd.build())
}

// registerSocketTools registers socket and connection tools
func (s *MCPServer) registerSocketTools(server *mcp.Server) {
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "ss",
			Description: `Show socket statistics and connections.

Examples:
- Show all TCP connections: {"protocol": "tcp"}
- Show listening sockets: {"state": "listening"}
- Show with process info: {"protocol": "tcp", "process": true}
- Filter by port: {"protocol": "tcp", "port_filter": "sport = :8080"}`,
		}, s.SS)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "netstat",
			Description: `Show network statistics (fallback for systems without ss).

Examples:
- Show all connections: {}
- Show listening TCP: {"protocol": "tcp", "listening": true}
- Numeric output: {"numeric": true}`,
		}, s.Netstat)
}
