package external_tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	kubernetesmcp "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/mcp"
)

// MCPServer provides OVS layer analysis tools
type MCPServer struct {
	k8sMcpServer *kubernetesmcp.MCPServer
}

// NewMCPServer creates a new OVS MCP server
func NewMCPServer(k8sMcpServer *kubernetesmcp.MCPServer) (*MCPServer, error) {
	return &MCPServer{
		k8sMcpServer: k8sMcpServer,
	}, nil
}

// AddTools registers external tools with the MCP server
func (s *MCPServer) AddTools(server *mcp.Server) {
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "tcpdump",
			Description: `Capture network packets with strict safety controls.

Parameters:
- Must specify at least 2 of: duration, packet_count, bpf_filter
- Maximum duration: 30 seconds
- Maximum packet_count: 1000
- Default snaplen: 96 bytes (headers only)
- For 'any' interface: requires BPF filter OR very low limits (duration<=10s, count<=100)

Examples:
- Capture HTTP traffic: {"interface": "eth0", "duration": 10, "bpf_filter": "tcp port 80"}
- Capture specific host: {"interface": "eth0", "packet_count": 100, "bpf_filter": "host 10.0.0.1"}
- Capture DNS: {"interface": "any", "duration": 5, "packet_count": 50, "bpf_filter": "port 53"}
- Full packet capture: {"interface": "eth0", "duration": 5, "bpf_filter": "host 192.168.1.1", "snaplen": 262}`,
		}, s.Tcpdump)
}
