package external_tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
)

const (
	MaxPacketCount = 1000
	DefaultSnaplen = 96
	MaxSnaplen     = 262
)

// Tcpdump executes the tcpdump packet capture tool on a node or inside a pod.
// It supports BPF filtering, snapshot length control, and configurable packet counts.
// Output can be in text format for immediate analysis or pcap format for offline analysis.
func (s *MCPServer) Tcpdump(ctx context.Context, req *mcp.CallToolRequest, in types.TcpdumpParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := ValidateInterface(in.Interface); err != nil {
		return nil, types.CommandResult{}, err
	}
	if err := ValidateBPFFilter(in.BPFFilter); err != nil {
		return nil, types.CommandResult{}, err
	}

	packetCount := in.PacketCount
	if packetCount == 0 {
		packetCount = MaxPacketCount
	}
	if err := validateIntMax(packetCount, MaxPacketCount, "packet_count", ""); err != nil {
		return nil, types.CommandResult{}, err
	}

	snaplen := in.Snaplen
	if snaplen == 0 {
		snaplen = DefaultSnaplen
	}
	if err := validateIntMax(snaplen, MaxSnaplen, "snaplen", "bytes"); err != nil {
		return nil, types.CommandResult{}, err
	}

	cmd := newCommand("tcpdump", "-n",
		"-s", strconv.Itoa(snaplen),
		"-c", strconv.Itoa(packetCount))
	cmd.addIfNotEmpty(in.Interface, "-i", in.Interface)
	cmd.addIfNotEmpty(in.BPFFilter, in.BPFFilter)

	outputFormat := stringWithDefault(in.OutputFormat, "text")
	switch outputFormat {
	case "text":
		cmd.add("-v")
	case "pcap":
		cmd.add("-w", "-")
	default:
		return nil, types.CommandResult{}, fmt.Errorf("invalid output_format: %s (must be 'text' or 'pcap')", outputFormat)
	}

	switch in.TargetType {
	case "node":
		_, result, err := s.runDebugNode(ctx, req, types.DebugNodeParams{NodeName: in.NodeName, NodeImage: in.NodeImage}, cmd.build())
		if err != nil {
			return nil, types.CommandResult{}, err
		}
		return nil, result, nil
	case "pod":
		_, result, err := s.runExecPod(ctx, req, types.ExecPodParams{PodName: in.PodName, PodNamespace: in.PodNamespace, ContainerName: in.ContainerName}, cmd.build())
		if err != nil {
			return nil, types.CommandResult{}, err
		}
		return nil, result, nil
	default:
		return nil, types.CommandResult{}, fmt.Errorf("invalid target_type: %s (must be 'node' or 'pod')", in.TargetType)

	}
}
