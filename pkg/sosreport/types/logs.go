// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/headtail"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/pattern"
)

// GetPodLogsParams are the parameters for sos-get-pod-logs
type GetPodLogsParams struct {
	SosreportPath string `json:"sosreport_path"`
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	Container     string `json:"container,omitempty"`
	pattern.PatternParams
	headtail.HeadTailParams
}

// GetPodLogsResult returns matching pod log lines
type GetPodLogsResult struct {
	Output string `json:"output"`
}
