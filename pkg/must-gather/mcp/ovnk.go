package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/types"
)

// GetOvnKInfo uses the omc command to get the ovnk info. It will return an error if the
// must gather path is not valid or the ovnk info is not found.
func (s *MustGatherMCPServer) GetOvnKInfo(ctx context.Context, req *mcp.CallToolRequest, in types.GetOvnKInfoParams) (*mcp.CallToolResult, types.GetOvnKInfoResult, error) {
	output, err := s.omcClient.GetOvnKInfo(ctx, in.MustGatherPath, in.InfoType)
	if err != nil {
		return nil, types.GetOvnKInfoResult{}, err
	}
	return nil, types.GetOvnKInfoResult{Data: output}, nil
}
