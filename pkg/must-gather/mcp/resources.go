package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/types"
)

// GetResource uses the omc command to get a resource. It will return an error if the
// must gather path is not valid or the resource is not found.
func (s *MustGatherMCPServer) GetResource(ctx context.Context, req *mcp.CallToolRequest, in types.GetResourceParams) (*mcp.CallToolResult, types.ResourceResult, error) {
	output, err := s.OmcClient.GetResource(in.MustGatherPath, in.Kind, in.Namespace, in.Name, in.OutputType)
	if err != nil {
		return nil, types.ResourceResult{}, err
	}
	return nil, types.ResourceResult{Data: output}, nil
}

// ListResources uses the omc command to list resources. It will return an error if the
// must gather path is not valid or the resources are not found.
func (s *MustGatherMCPServer) ListResources(ctx context.Context, req *mcp.CallToolRequest, in types.ListResourcesParams) (*mcp.CallToolResult, types.ResourceResult, error) {
	output, err := s.OmcClient.ListResources(in.MustGatherPath, in.Kind, in.Namespace, in.LabelSelector, in.OutputType)
	if err != nil {
		return nil, types.ResourceResult{}, err
	}
	return nil, types.ResourceResult{Data: output}, nil
}
