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
			Description: `Capture network packets on a node or inside a pod with strict safety controls.

Supports both node-level and pod-level packet capture with BPF filtering.

This tool creates a specialized debug pod on the specified node with the nicolaka/netshoot:v0.13 image
for node-level captures, ensuring consistent and secure packet capture capabilities.

Parameters:
- target_type: 'node' or 'pod' (required)
- node_name: Name of the node (required when target_type is 'node')
- pod_name: Name of the pod (required when target_type is 'pod')
- pod_namespace: Namespace of the pod (required when target_type is 'pod')
- container_name: Name of the container in the pod (optional, uses default container if not specified)
- interface: Network interface name or 'any' (optional, uses default if not specified)
- packet_count: Number of packets to capture (default: 100, max: 1000)
- bpf_filter: BPF filter expression (optional)
- snaplen: Snapshot length in bytes (default: 96, max: 1500)
- output_format: Output format 'text' or 'pcap' (default: 'text')

Examples:
- Capture on node: {"target_type": "node", "node_name": "worker-1", "interface": "eth0", "packet_count": 100, "bpf_filter": "tcp port 80"}
- Capture in pod: {"target_type": "pod", "pod_name": "my-pod", "pod_namespace": "default", "interface": "eth0", "packet_count": 100, "bpf_filter": "host 10.0.0.1"}
- Capture DNS: {"target_type": "node", "node_name": "worker-1", "interface": "any", "packet_count": 50, "bpf_filter": "port 53"}`,
		}, s.Tcpdump)
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "pwru",
			Description: `Trace packets through the Linux kernel networking stack using eBPF.

pwru (packet, where are you?) shows which kernel functions process a packet, helping debug packet
drops, routing issues, and understanding the kernel's packet processing path.

This tool creates a specialized debug pod on the specified node with the docker.io/cilium/pwru:v1.0.10 image
and necessary eBPF capabilities to trace packets through kernel networking functions.

Parameters:
- node_name: Name of the node to run pwru on (required)
- pcap_filter: PCAP filter expression to match packets (optional, e.g., "tcp and dst port 8080", "host 10.0.0.1")
- output_limit_lines: Maximum number of trace events to capture (default: 100, max: 1000)

Examples:
- Basic trace: {"node_name": "worker-1", "pcap_filter": "host 10.244.0.5", "output_limit_lines": 100}
- TCP traffic: {"node_name": "worker-1", "pcap_filter": "tcp and dst port 8080", "output_limit_lines": 50}
- ICMP packets: {"node_name": "worker-1", "pcap_filter": "icmp", "output_limit_lines": 100}`,
		}, s.Pwru)
}
