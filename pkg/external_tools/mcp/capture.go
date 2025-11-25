package external_tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/client"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
)

const (
	MaxTcpdumpDuration = 30
	MaxPacketCount     = 1000
	DefaultSnaplen     = 96
	MaxSnaplen         = 262
)

func (s *MCPServer) Tcpdump(ctx context.Context, req *mcp.CallToolRequest, in types.TcpdumpParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := client.ValidateInterface(in.Interface); err != nil {
		return nil, types.CommandResult{}, err
	}
	if err := client.ValidateBPFFilter(in.BPFFilter); err != nil {
		return nil, types.CommandResult{}, err
	}

	if in.Interface == "any" {
		if in.BPFFilter == "" {
			if in.Duration == 0 && in.PacketCount == 0 {
				return nil, types.CommandResult{}, fmt.Errorf("capturing on 'any' interface requires either a BPF filter or explicit duration/packet_count limits")
			}
			if in.Duration > 10 || in.PacketCount > 100 {
				return nil, types.CommandResult{}, fmt.Errorf("capturing on 'any' interface without BPF filter requires duration <= 10s and packet_count <= 100")
			}
		}
	}

	if err := requireAtLeastNParams(2, map[string]bool{
		"duration":     in.Duration > 0,
		"packet_count": in.PacketCount > 0,
		"bpf_filter":   in.BPFFilter != "",
	}); err != nil {
		return nil, types.CommandResult{}, fmt.Errorf("tcpdump %w", err)
	}

	if err := validateIntMax(in.Duration, MaxTcpdumpDuration, "duration", "seconds"); err != nil {
		return nil, types.CommandResult{}, err
	}

	if err := validateIntMax(in.PacketCount, MaxPacketCount, "packet_count", ""); err != nil {
		return nil, types.CommandResult{}, err
	}

	snaplen := in.Snaplen
	if snaplen == 0 {
		snaplen = DefaultSnaplen
	}
	if err := validateIntMax(snaplen, MaxSnaplen, "snaplen", "bytes"); err != nil {
		return nil, types.CommandResult{}, err
	}

	cmd := newCommand("tcpdump",
		"-i", in.Interface,
		"-n",
		"-s", strconv.Itoa(snaplen)).
		addIf(in.PacketCount > 0, "-c", strconv.Itoa(in.PacketCount))

	outputFormat := stringWithDefault(in.OutputFormat, "text")
	switch outputFormat {
	case "text":
		cmd.add("-v")
	case "pcap":
		cmd.add("-w", "-")
	default:
		return nil, types.CommandResult{}, fmt.Errorf("invalid output_format: %s (must be 'text' or 'pcap')", outputFormat)
	}
	cmd.addIfNotEmpty(in.BPFFilter, in.BPFFilter)
	return s.executeAndWrapResult(ctx, in.TargetParams, cmd.build())
}

// registerCaptureTools registers packet capture tools
func (s *MCPServer) registerCaptureTools(server *mcp.Server) {
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "tcpdump",
			Description: `Capture network packets with strict safety controls.

IMPORTANT SAFETY REQUIREMENTS:
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
