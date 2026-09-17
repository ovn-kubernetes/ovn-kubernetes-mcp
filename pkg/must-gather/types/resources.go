// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package types

import k8sTypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"

// GetResourceParams is a type that contains the must gather path, kind and get params.
type GetResourceParams struct {
	MustGatherParams
	Kind string `json:"kind"`
	k8sTypes.GetParams
}

// ListResourcesParams is a type that contains the must gather path, kind and list params.
type ListResourcesParams struct {
	MustGatherParams
	Kind string `json:"kind"`
	k8sTypes.ListParams
}

// ResourceResult is a type that contains the resource data.
type ResourceResult struct {
	Data string `json:"data"`
}
