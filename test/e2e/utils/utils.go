// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/inspector"
)

func UnmarshalCallToolResult[T any](output []byte) T {
	var result mcp.CallToolResult
	err := result.UnmarshalJSON(output)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(result.StructuredContent).NotTo(gomega.BeEmpty())

	jsonOutput, err := json.Marshal(result.StructuredContent)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	var resultInTFormat T
	err = json.Unmarshal(jsonOutput, &resultInTFormat)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return resultInTFormat
}

func ExpectToolError(mcpInspector *inspector.MCPInspector, toolName string, toolArgs map[string]any, wantError string) {
	output, err := mcpInspector.
		MethodCall(toolName, toolArgs).
		Execute()

	if err != nil {
		gomega.Expect(output).NotTo(gomega.BeEmpty(),
			"tool error returned exit code but stdout is empty: %v", err)
	}

	var result mcp.CallToolResult
	gomega.Expect(result.UnmarshalJSON(output)).To(gomega.Succeed())
	gomega.Expect(result.IsError).To(gomega.BeTrue())
	gomega.Expect(result.Content).NotTo(gomega.BeEmpty())

	var messages []string
	for _, content := range result.Content {
		textContent, ok := content.(*mcp.TextContent)
		if ok {
			messages = append(messages, textContent.Text)
		}
	}

	gomega.Expect(messages).NotTo(gomega.BeEmpty(), "expected error text content in CallToolResult")
	gomega.Expect(strings.Join(messages, "\n")).To(gomega.ContainSubstring(wantError))
}

// RestrictedSecurityContext returns a SecurityContext that complies with the
// "restricted" PodSecurity standard (no privilege escalation, drops all caps,
// runs as non-root user 1000, RuntimeDefault seccomp profile).
func RestrictedSecurityContext() *corev1.SecurityContext {
	allowPrivilegeEscalation := false
	runAsNonRoot := true
	runAsUser := int64(1000)
	return &corev1.SecurityContext{
		AllowPrivilegeEscalation: &allowPrivilegeEscalation,
		RunAsNonRoot:             &runAsNonRoot,
		RunAsUser:                &runAsUser,
		Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
		SeccompProfile:           &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
	}
}

func GetTestdataPath(relativePath string) string {
	_, thisFile, _, ok := runtime.Caller(1)
	gomega.Expect(ok).To(gomega.BeTrue())

	path, err := filepath.Abs(filepath.Join(filepath.Dir(thisFile), relativePath))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	_, err = os.Stat(path)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return path
}
