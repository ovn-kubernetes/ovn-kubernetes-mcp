package external_tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
)

const (
	MaxTcpdumpDuration = 30
	MaxPacketCount     = 1000
	DefaultSnaplen     = 96
	MaxSnaplen         = 262
)

func (s *MCPServer) Tcpdump(ctx context.Context, req *mcp.CallToolRequest, in types.TcpdumpParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := ValidateInterface(in.Interface); err != nil {
		return nil, types.CommandResult{}, err
	}
	if err := ValidateBPFFilter(in.BPFFilter); err != nil {
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
