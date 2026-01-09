package external_tools

import (
	"context"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
	k8stypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
)

const (
	DefaultOutputLimitLines = 100
	MaxOutputLimitLines     = 1000
)

// Pwru executes the pwru (packet, where are you?) tool to trace packets through the Linux kernel.
// It creates a specialized debug pod with eBPF capabilities and traces packet processing paths.
// This is useful for debugging packet drops, routing issues, and understanding kernel networking behavior.
func (s *MCPServer) Pwru(ctx context.Context, req *mcp.CallToolRequest, in types.PwruParams) (*mcp.CallToolResult, types.CommandResult, error) {
	outputLimitLines := in.OutputLimitLines
	if outputLimitLines == 0 {
		outputLimitLines = DefaultOutputLimitLines
	}
	if err := validateIntMax(outputLimitLines, MaxOutputLimitLines, "output_limit_lines", ""); err != nil {
		return nil, types.CommandResult{}, err
	}

	if err := validatePacketFilter(in.PcapFilter); err != nil {
		return nil, types.CommandResult{}, err
	}

	cmd := newCommand("pwru", "--output-limit-lines", strconv.Itoa(outputLimitLines))
	cmd.addIfNotEmpty(in.PcapFilter, in.PcapFilter)

	target := k8stypes.DebugNodeParams{
		Name:      in.NodeName,
		Image:     "docker.io/cilium/pwru:v1.0.10",
		HostPath:  "/sys/kernel/debug",
		MountPath: "/sys/kernel/debug",
		Command:   cmd.build(),
	}

	result, err := s.runDebugNode(ctx, req, target)
	if err != nil {
		return nil, types.CommandResult{}, err
	}
	return nil, result, nil
}
