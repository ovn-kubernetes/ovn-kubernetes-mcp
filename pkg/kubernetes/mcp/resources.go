package mcp

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
)

// GetResource gets a resource by group, version, kind, name and namespace.
func (s *MCPServer) GetResource(ctx context.Context, req *mcp.CallToolRequest, in types.GetResourceParams) (*mcp.CallToolResult, types.GetResourceResult, error) {
	// If the version, kind or name is not set, return an error.
	var err error
	if in.Version == "" {
		err = errors.New("version is required")
	}
	if in.Kind == "" {
		err = errors.Join(err, errors.New("kind is required"))
	}
	if in.Name == "" {
		err = errors.Join(err, errors.New("name is required"))
	}
	if err != nil {
		return nil, types.GetResourceResult{}, err
	}

	// Get the resource by group, version, kind, name and namespace.
	resource, err := s.clientSet.GetResource(in.Group, in.Version, in.Kind, in.Name, in.Namespace)
	if err != nil {
		return nil, types.GetResourceResult{}, err
	}

	resourceData := types.Resource{}
	// Get the formatted data from the resource.
	switch in.OutputType {
	case types.JSONOutputType:
		err = resourceData.ToJSON(resource)
		if err != nil {
			return nil, types.GetResourceResult{}, err
		}
	case types.YAMLOutputType:
		err = resourceData.ToYAML(resource)
		if err != nil {
			return nil, types.GetResourceResult{}, err
		}
	default:
		// If the output type is not JSON or YAML, get the resource data.
		resourceData.GetResourceData(resource, in.OutputType == types.WideOutputType)
	}

	return nil, types.GetResourceResult{Resource: resourceData}, nil
}

// ListResources lists resources by group, version, kind and namespace.
func (s *MCPServer) ListResources(ctx context.Context, req *mcp.CallToolRequest, in types.ListResourcesParams) (*mcp.CallToolResult, types.ListResourcesResult, error) {
	// If the version or kind is not set, return an error.
	var err error
	if in.Version == "" {
		err = errors.New("version is required")
	}
	if in.Kind == "" {
		err = errors.Join(err, errors.New("kind is required"))
	}
	if err != nil {
		return nil, types.ListResourcesResult{}, err
	}

	// List the resources by group, version, kind and namespace.
	resources, err := s.clientSet.ListResources(in.Group, in.Version, in.Kind, in.Namespace, in.LabelSelector)
	if err != nil {
		return nil, types.ListResourcesResult{}, err
	}

	// If there are no resources, return an empty list.
	if len(resources.Items) == 0 {
		return nil, types.ListResourcesResult{Resources: []types.Resource{}}, nil
	}

	resourcesData := make([]types.Resource, 0)
	// Loop through the resources and get the resource data.
	for _, resource := range resources.Items {
		resourceData := types.Resource{}
		// Get the formatted data from the resource.
		switch in.OutputType {
		case types.JSONOutputType:
			err = resourceData.ToJSON(&resource)
			if err != nil {
				return nil, types.ListResourcesResult{}, err
			}
		case types.YAMLOutputType:
			err = resourceData.ToYAML(&resource)
			if err != nil {
				return nil, types.ListResourcesResult{}, err
			}
		default:
			// If the output type is not JSON or YAML, get the resource data.
			resourceData.GetResourceData(&resource, in.OutputType == types.WideOutputType)
		}
		resourcesData = append(resourcesData, resourceData)
	}

	return nil, types.ListResourcesResult{Resources: resourcesData}, nil
}
