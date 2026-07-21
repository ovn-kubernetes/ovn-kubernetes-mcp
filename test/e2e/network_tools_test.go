package e2e

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/network-tools/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("Network Tools", func() {
	const (
		tcpdumpToolName = "tcpdump"
		pwruToolName    = "pwru"
	)

	var nodeName string

	extractToolError := func(callResult mcp.CallToolResult) string {
		var messages []string
		for _, c := range callResult.Content {
			if tc, ok := c.(*mcp.TextContent); ok {
				messages = append(messages, tc.Text)
			}
		}
		if len(messages) > 0 {
			return fmt.Sprintf("tool error: %s", strings.Join(messages, "; "))
		}
		return "tool error: unknown (no text content)"
	}

	// BeforeEach finds a ready worker node once per test
	BeforeEach(func() {
		nodeName = ""
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		nodeList := &corev1.NodeList{}
		err := kubeClient.List(ctx, nodeList)
		Expect(err).NotTo(HaveOccurred(), "Failed to list nodes")

		// Look for a ready worker node (not control-plane)
		for _, node := range nodeList.Items {
			isReady := false
			isControlPlane := false

			for _, condition := range node.Status.Conditions {
				if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
					isReady = true
					break
				}
			}

			if _, hasLabel := node.Labels["node-role.kubernetes.io/control-plane"]; hasLabel {
				isControlPlane = true
			}
			if _, hasLabel := node.Labels["node-role.kubernetes.io/master"]; hasLabel {
				isControlPlane = true
			}

			if isReady && !isControlPlane {
				nodeName = node.Name
				break
			}
		}

		if nodeName == "" {
			// If no worker nodes found, use any ready node
			for _, node := range nodeList.Items {
				isReady := false
				for _, condition := range node.Status.Conditions {
					if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
						isReady = true
						break
					}
				}
				if isReady {
					nodeName = node.Name
					break
				}
			}
		}

		Expect(nodeName).NotTo(BeEmpty(), "No ready nodes found")
	})

	callTcpdump := func(args map[string]any) types.CommandResult {
		args["timeout_seconds"] = 30
		var result types.CommandResult
		Eventually(func() bool {
			output, err := mcpInspector.
				MethodCall(tcpdumpToolName, args).Execute()
			if err != nil {
				GinkgoWriter.Printf("tcpdump retry: execute error: %v\n", err)
				return false
			}
			if len(output) == 0 {
				GinkgoWriter.Println("tcpdump retry: empty output")
				return false
			}
			var callResult mcp.CallToolResult
			if err := callResult.UnmarshalJSON(output); err != nil {
				GinkgoWriter.Printf("tcpdump retry: unmarshal error: %v\n", err)
				return false
			}
			if callResult.IsError {
				GinkgoWriter.Printf("tcpdump retry: %s\n", extractToolError(callResult))
				return false
			}
			result = utils.UnmarshalCallToolResult[types.CommandResult](output)
			return result.Output != ""
		}, 120*time.Second, 10*time.Second).Should(BeTrue(),
			"tcpdump never returned successfully; check test output for error details")
		return result
	}

	Context("tcpdump", func() {
		It("should capture network packets with minimal parameters", func() {
			By("Running tcpdump with basic node capture")
			result := callTcpdump(map[string]any{
				"target_type":  "node",
				"name":         nodeName,
				"packet_count": 2,
				"interface":    "any",
			})

			By("Verifying tcpdump output contains network traffic")
			Expect(result.Output).To(Or(
				ContainSubstring("IP "),
				ContainSubstring("TCP "),
				ContainSubstring("UDP "),
				ContainSubstring(" > "),
			))
		})

		It("should capture network packets with interface 'any'", func() {
			By("Running tcpdump with any interface")
			result := callTcpdump(map[string]any{
				"target_type":  "node",
				"name":         nodeName,
				"packet_count": 3,
				"interface":    "any",
			})

			By("Verifying tcpdump output with any interface")
			Expect(result.Output).To(Or(
				ContainSubstring("IP "),
				ContainSubstring("In "),
				ContainSubstring("Out "),
				ContainSubstring("lo "),
			))
		})

		It("should capture network packets with BPF filter", func() {
			By("Running tcpdump with TCP filter")
			result := callTcpdump(map[string]any{
				"target_type":  "node",
				"name":         nodeName,
				"packet_count": 5,
				"interface":    "any",
				"bpf_filter":   "tcp",
			})

			By("Verifying tcpdump BPF filter output")
			Expect(result.Output).NotTo(BeEmpty())
			Expect(strings.ToLower(result.Output)).To(ContainSubstring("tcp"))
		})

		It("should handle custom snapshot length", func() {
			By("Running tcpdump with custom snaplen")
			result := callTcpdump(map[string]any{
				"target_type":  "node",
				"name":         nodeName,
				"packet_count": 2,
				"interface":    "any",
				"snaplen":      128,
			})

			By("Verifying tcpdump custom snaplen output")
			Expect(result.Output).NotTo(BeEmpty())
			Expect(result.Output).To(ContainSubstring("IP "))
		})
	})

	Context("pwru", func() {
		It("should trace kernel network stack", func() {
			if os.Getenv("CI") == "true" {
				Skip("pwru requires BPF/BTF kernel support unavailable in Kind containers")
			}

			By("Running pwru to trace kernel networking")
			output, err := mcpInspector.
				MethodCall(pwruToolName, map[string]any{
					"node_name":          nodeName,
					"output_limit_lines": 10,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Verifying pwru kernel trace output")
			result := utils.UnmarshalCallToolResult[types.CommandResult](output)
			Expect(result.Output).NotTo(BeEmpty())
			Expect(result.Output).To(Or(
				ContainSubstring("SKB"),
				ContainSubstring("CPU"),
				ContainSubstring("PROCESS"),
				ContainSubstring("NETNS"),
				ContainSubstring("FUNC"),
			))
		})

		It("should trace with BPF filter", func() {
			if os.Getenv("CI") == "true" {
				Skip("pwru requires BPF/BTF kernel support unavailable in Kind containers")
			}

			By("Running pwru with host filter")
			output, err := mcpInspector.
				MethodCall(pwruToolName, map[string]any{
					"node_name":          nodeName,
					"bpf_filter":         "tcp",
					"output_limit_lines": 8,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Verifying pwru filtered trace output")
			result := utils.UnmarshalCallToolResult[types.CommandResult](output)
			Expect(result.Output).NotTo(BeEmpty())
			Expect(result.Output).To(Or(
				ContainSubstring("SKB"),
				ContainSubstring("FUNC"),
				ContainSubstring("CPU"),
			))
		})

		It("should limit output lines correctly", func() {
			if os.Getenv("CI") == "true" {
				Skip("pwru requires BPF/BTF kernel support unavailable in Kind containers")
			}

			By("Running pwru with small line limit")
			output, err := mcpInspector.
				MethodCall(pwruToolName, map[string]any{
					"node_name":          nodeName,
					"output_limit_lines": 3,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Verifying pwru line limit output")
			result := utils.UnmarshalCallToolResult[types.CommandResult](output)
			Expect(result.Output).NotTo(BeEmpty())
			lines := strings.Split(strings.TrimSpace(result.Output), "\n")
			Expect(len(lines)).To(BeNumerically(">=", 1), "Should have at least header line")
			Expect(len(lines)).To(BeNumerically("<=", 4), "Should respect line limit parameter")
		})
	})

})
