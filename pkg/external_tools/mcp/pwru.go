package external_tools

import (
	"context"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/types"
)

const (
	MaxOutputLimitLines = 1000
)

// Pwru executes the pwru (packet, where are you?) tool to trace packets through the Linux kernel.
// It creates a specialized debug pod with eBPF capabilities and traces packet processing paths.
// This is useful for debugging packet drops, routing issues, and understanding kernel networking behavior.
func (s *MCPServer) Pwru(ctx context.Context, req *mcp.CallToolRequest, in types.PwruParams) (*mcp.CallToolResult, types.CommandResult, error) {
	outputLimitLines := in.OutputLimitLines
	if outputLimitLines == 0 {
		outputLimitLines = MaxOutputLimitLines
	}
	if err := validateIntMax(outputLimitLines, MaxOutputLimitLines, "output_limit_lines", ""); err != nil {
		return nil, types.CommandResult{}, err
	}

	if err := ValidateBPFFilter(in.PcapFilter); err != nil {
		return nil, types.CommandResult{}, err
	}

	cmd := newCommand("pwru", "--output-limit-lines", strconv.Itoa(outputLimitLines))
	cmd.addIfNotEmpty(in.PcapFilter, in.PcapFilter)

	hostPath := "/sys/kernel/debug"
	mountPath := "/sys/kernel/debug"

	target := types.DebugNodeParams{
		NodeName:  in.NodeName,
		NodeImage: "docker.io/cilium/pwru:v1.0.10",
		HostPath:  hostPath,
		MountPath: mountPath,
	}

	_, result, err := s.runDebugNode(ctx, req, target, cmd.build())
	if err != nil {
		return nil, types.CommandResult{}, err
	}
	return nil, result, nil
}
