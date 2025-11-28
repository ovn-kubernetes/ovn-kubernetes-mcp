package nettools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	k8stypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/nettools/types"
)

const (
	DefaultPacketCount = 100
	MaxPacketCount     = 1000
	// DefaultSnaplen is set to 96 bytes to capture headers (Ethernet + IP + TCP/UDP + some payload)
	// This is sufficient for most diagnostic purposes while minimizing data capture
	DefaultSnaplen = 96
	// MaxSnaplen is set to 1500 bytes (standard Ethernet MTU) to allow full packet capture
	// when needed for deeper analysis. Users can set snaplen parameter to capture complete packets.
	MaxSnaplen = 1500
)

// Tcpdump executes the tcpdump packet capture tool on a node or inside a pod.
func (s *MCPServer) Tcpdump(ctx context.Context, req *mcp.CallToolRequest, in types.TcpdumpParams) (*mcp.CallToolResult, types.CommandResult, error) {
	if err := validateInterface(in.Interface); err != nil {
		return nil, types.CommandResult{}, err
	}
	if err := validatePacketFilter(in.BPFFilter); err != nil {
		return nil, types.CommandResult{}, err
	}

	packetCount := in.PacketCount
	if packetCount == 0 {
		packetCount = DefaultPacketCount
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

	cmd := newCommand("tcpdump", "-n", "-v",
		"-s", strconv.Itoa(snaplen),
		"-c", strconv.Itoa(packetCount))
	cmd.addIfNotEmpty(in.Interface, "-i", in.Interface)
	cmd.addIfNotEmpty(in.BPFFilter, in.BPFFilter)

	switch in.TargetType {
	case "node":
		result, err := s.runDebugNode(ctx, req, k8stypes.DebugNodeParams{
			Name:    in.NodeName,
			Image:   s.tcpdumpImage,
			Command: cmd.build(),
		})
		if err != nil {
			return nil, types.CommandResult{}, err
		}
		return nil, result, nil
	case "pod":
		result, err := s.runExecPod(ctx, req, k8stypes.ExecPodParams{
			NamespacedNameParams: k8stypes.NamespacedNameParams{
				Name:      in.PodName,
				Namespace: in.PodNamespace,
			},
			Container: in.ContainerName,
			Command:   cmd.build(),
		})
		if err != nil {
			return nil, types.CommandResult{}, err
		}
		return nil, result, nil
	default:
		return nil, types.CommandResult{}, fmt.Errorf("invalid target_type: %s (must be 'node' or 'pod')", in.TargetType)
	}
}
