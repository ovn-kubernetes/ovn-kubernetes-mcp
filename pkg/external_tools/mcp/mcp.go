package external_tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	kubernetesmcp "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/mcp"
)

// MCPServer provides MCP server functionality for external tools operations.
type MCPServer struct {
	k8sMcpServer *kubernetesmcp.MCPServer
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(k8sMcpServer *kubernetesmcp.MCPServer) *MCPServer {
	return &MCPServer{
		k8sMcpServer: k8sMcpServer,
	}
}

// AddTools registers external tools with the MCP server
func (s *MCPServer) AddTools(server *mcp.Server) {
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "tcpdump",
			Description: `Capture network packets with strict safety controls.

Parameters:
- packet_count: Number of packets to capture (default: 1000, max: 1000)
- bpf_filter: BPF filter expression (optional)
- snaplen: Snapshot length in bytes (default: 96, max: 262)
- interface: Network interface name or 'any' (optional, uses default if not specified)
- target_type: 'node' or 'pod' (optional, uses default interface if not specified)

Examples:
- Capture HTTP traffic: {"interface": "eth0", "packet_count": 100, "bpf_filter": "tcp port 80"}
- Capture specific host: {"interface": "eth0", "packet_count": 100, "bpf_filter": "host 10.0.0.1"}
- Capture DNS: {"interface": "any", "packet_count": 50, "bpf_filter": "port 53"}
- Full packet capture: {"interface": "eth0", "packet_count": 100, "bpf_filter": "host 192.168.1.1", "snaplen": 262}`,
		}, s.Tcpdump)
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "pwru",
			Description: `Trace packets through the Linux kernel networking stack using eBPF.

pwru (packet, where are you?) shows which kernel functions process a packet, helping debug packet
drops, routing issues, and understanding the kernel's packet processing path.

This tool uses a specialized debug pod with the cilium/pwru:v1.0.10 image and necessary eBPF capabilities.

Parameters:
- node_name: Name of the node to run pwru on (required)
- pcap_filter: pcap expression to filter packets (e.g., "tcp and dst port 8080", "host 10.0.0.1")
- output_limit_lines: Stop after N events (default: 1000, max: 1000)

Examples:
- Basic trace: {"node_name": "worker-1", "pcap_filter": "host 10.244.0.5", "output_limit_lines": 100}
- TCP traffic: {"node_name": "worker-1", "pcap_filter": "tcp and dst port 8080", "output_limit_lines": 50}
- ICMP packets: {"node_name": "worker-1", "pcap_filter": "icmp", "output_limit_lines": 100}`,
		}, s.Pwru)
}
