package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/types"
)

// GetResource uses the omc command to get a resource. It will return an error if the
// must gather path is not valid or the resource is not found.
func (s *MustGatherMCPServer) GetResource(ctx context.Context, req *mcp.CallToolRequest, in types.GetResourceParams) (*mcp.CallToolResult, types.ResourceResult, error) {
	// If the kind or name is not set, return an error.
	var err error
	if in.Kind == "" {
		err = errors.New("kind is required")
	}
	if in.Name == "" {
		err = errors.Join(err, errors.New("name is required"))
	}
	if err != nil {
		return nil, types.ResourceResult{}, err
	}

	output, err := s.omcClient.GetResource(ctx, in.MustGatherPath, in.Kind, in.Namespace, in.Name, in.OutputType)
	if err != nil {
		return nil, types.ResourceResult{}, err
	}
	return nil, types.ResourceResult{Data: output}, nil
}

// ListResources uses the omc command to list resources. It will return an error if the
// must gather path is not valid.
func (s *MustGatherMCPServer) ListResources(ctx context.Context, req *mcp.CallToolRequest, in types.ListResourcesParams) (*mcp.CallToolResult, types.ResourceResult, error) {
	// If the kind is not set, return an error.
	if in.Kind == "" {
		return nil, types.ResourceResult{}, fmt.Errorf("kind is required")
	}

	output, err := s.omcClient.ListResources(ctx, in.MustGatherPath, in.Kind, in.Namespace, in.LabelSelector, in.OutputType)
	if err != nil {
		return nil, types.ResourceResult{}, err
	}
	return nil, types.ResourceResult{Data: output}, nil
}
