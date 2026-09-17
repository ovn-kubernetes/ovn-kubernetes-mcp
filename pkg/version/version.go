// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package version

import "fmt"

const (
	Version = "0.1.0"
)

func Print() string {
	return fmt.Sprintf("ovn-kubernetes-mcp-server version %s", Version)
}
