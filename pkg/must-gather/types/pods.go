package types

import k8sTypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"

// GetPodLogsParams is a type that contains the must gather path, get pod logs params,
// rotated, pattern, head and tail.
type GetPodLogsParams struct {
	MustGatherParams
	k8sTypes.GetPodLogsParams
	Rotated bool   `json:"rotated,omitempty"`
	Pattern string `json:"pattern,omitempty"`
	Head    int    `json:"head,omitempty"`
	Tail    int    `json:"tail,omitempty"`
}

// GetPodLogsResult is a type that contains the logs of a pod where each log line
// is a separate element in the string slice.
type GetPodLogsResult struct {
	Logs []string `json:"logs"`
}
