package e2e

import (
	"context"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kernel/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("Kernel Tools", func() {
	const (
		getIPToolName        = "get-ip"
		getIPTablesToolName  = "get-iptables"
		getNFTToolName       = "get-nft"
		getConntrackToolName = "get-conntrack"
	)

	var nodeName string

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
				return
			}
		}

		// If no worker found, use the first ready node
		for _, node := range nodeList.Items {
			for _, condition := range node.Status.Conditions {
				if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
					nodeName = node.Name
					return
				}
			}
		}

		Expect(nodeName).NotTo(BeEmpty(), "No ready node found")
	})

	withNode := func(args map[string]any) map[string]any {
		toolArgs := make(map[string]any, len(args)+1)
		for key, value := range args {
			toolArgs[key] = value
		}
		toolArgs["node"] = nodeName
		return toolArgs
	}

	expectToolError := func(toolName string, toolArgs map[string]any, wantError string) {
		output, err := mcpInspector.
			MethodCall(toolName, toolArgs).
			Execute()
		Expect(err).NotTo(HaveOccurred())

		var result mcp.CallToolResult
		Expect(result.UnmarshalJSON(output)).To(Succeed())
		Expect(result.IsError).To(BeTrue())
		Expect(result.Content).NotTo(BeEmpty())

		var messages []string
		for _, content := range result.Content {
			textContent, ok := content.(*mcp.TextContent)
			if ok {
				messages = append(messages, textContent.Text)
			}
		}

		Expect(messages).NotTo(BeEmpty(), "expected error text content in CallToolResult")
		Expect(strings.Join(messages, "\n")).To(ContainSubstring(wantError))
	}

	Context("get-ip", func() {
		It("should retrieve IP routing information from a node", func() {
			By("Running get-ip to show routes")
			output, err := mcpInspector.
				MethodCall(getIPToolName, map[string]any{
					"node":    nodeName,
					"command": "route show",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains routing information")
			result := utils.UnmarshalCallToolResult[types.Result](output)
			Expect(result.Data).NotTo(BeEmpty())
			// Route output typically contains "via" or "dev" keywords
			Expect(result.Data).To(Or(
				ContainSubstring("via"),
				ContainSubstring("dev"),
			))
		})

		It("should retrieve network interface information from a node", func() {
			By("Running get-ip to show links")
			output, err := mcpInspector.
				MethodCall(getIPToolName, map[string]any{
					"node":    nodeName,
					"command": "link show",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains network interface information")
			result := utils.UnmarshalCallToolResult[types.Result](output)
			Expect(result.Data).NotTo(BeEmpty())
			// Link output typically contains interface names and states
			Expect(result.Data).To(Or(
				ContainSubstring("state"),
				ContainSubstring("mtu"),
				ContainSubstring("qdisc"),
			))
		})
	})

	Context("get-iptables", func() {
		It("should retrieve iptables rules from a node", func() {
			By("Running get-iptables to list filter table rules")
			output, err := mcpInspector.
				MethodCall(getIPTablesToolName, map[string]any{
					"node":    nodeName,
					"table":   "filter",
					"command": "-L",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains iptables rules")
			result := utils.UnmarshalCallToolResult[types.Result](output)
			Expect(result.Data).NotTo(BeEmpty())
			// iptables output typically contains chain names
			Expect(result.Data).To(Or(
				ContainSubstring("Chain"),
				ContainSubstring("target"),
				ContainSubstring("policy"),
			))
		})

		It("should retrieve NAT table rules from a node", func() {
			By("Running get-iptables to list nat table rules")
			output, err := mcpInspector.
				MethodCall(getIPTablesToolName, map[string]any{
					"node":    nodeName,
					"table":   "nat",
					"command": "-L",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains NAT rules")
			result := utils.UnmarshalCallToolResult[types.Result](output)
			Expect(result.Data).NotTo(BeEmpty())
			// NAT table output should contain PREROUTING, POSTROUTING chains
			Expect(result.Data).To(Or(
				ContainSubstring("PREROUTING"),
				ContainSubstring("POSTROUTING"),
				ContainSubstring("Chain"),
			))
		})
	})

	Context("get-nft", func() {
		It("should retrieve nftables ruleset from a node", func() {
			By("Running get-nft to list tables")
			output, err := mcpInspector.
				MethodCall(getNFTToolName, map[string]any{
					"node":    nodeName,
					"command": "list tables",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains nftables information")
			result := utils.UnmarshalCallToolResult[types.Result](output)
			Expect(result.Data).NotTo(BeEmpty(), "nftables should be configured on OVN-Kubernetes nodes")
			// nftables output typically contains "table"
			Expect(result.Data).To(ContainSubstring("table"))
		})
	})

	Context("validation errors", func() {
		DescribeTable("should reject invalid kernel tool inputs",
			func(toolName string, toolArgs map[string]any, wantError string) {
				expectToolError(toolName, withNode(toolArgs), wantError)
			},
			Entry("get-ip invalid command", getIPToolName, map[string]any{
				"command": "link list",
			}, "invalid ip command"),
			Entry("get-iptables invalid command", getIPTablesToolName, map[string]any{
				"table":   "filter",
				"command": "list",
			}, "invalid iptables command"),
			Entry("get-iptables invalid table", getIPTablesToolName, map[string]any{
				"table":   "invalid_table",
				"command": "-L",
			}, "invalid table name"),
			Entry("get-nft invalid command", getNFTToolName, map[string]any{
				"command": "list",
			}, "invalid nft command"),
			Entry("get-nft invalid address family", getNFTToolName, map[string]any{
				"command":          "list tables",
				"address_families": "invalid_family",
			}, "invalid nft address family"),
			Entry("get-conntrack invalid command", getConntrackToolName, map[string]any{
				"command": "list",
			}, "invalid command"),
		)

		DescribeTable("should reject metacharacters in filter parameters",
			func(toolName string, toolArgs map[string]any) {
				expectToolError(toolName, withNode(toolArgs), "invalid use of metacharacters in parameter")
			},
			Entry("get-ip filter parameters", getIPToolName, map[string]any{
				"command":           "route show",
				"filter_parameters": "table all; true",
			}),
			Entry("get-iptables filter parameters", getIPTablesToolName, map[string]any{
				"table":             "filter",
				"command":           "-L",
				"filter_parameters": "-nv | wc",
			}),
			Entry("get-nft address families", getNFTToolName, map[string]any{
				"command":          "list tables",
				"address_families": "inet; true",
			}),
			Entry("get-conntrack filter parameters", getConntrackToolName, map[string]any{
				"command":           "-L",
				"filter_parameters": "-s 1.2.3.4 $(true)",
			}),
		)
	})

	Context("get-conntrack", func() {
		It("should retrieve connection tracking count from a node", func() {
			By("Running get-conntrack to count connections")
			output, err := mcpInspector.
				MethodCall(getConntrackToolName, map[string]any{
					"node":    nodeName,
					"command": "-C",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains connection count")
			result := utils.UnmarshalCallToolResult[types.Result](output)
			Expect(result.Data).NotTo(BeEmpty())
			// Count output is just a number
			Expect(result.Data).To(MatchRegexp(`^\s*\d+\s*$`))
		})

		It("should retrieve connection tracking statistics from a node", func() {
			By("Running get-conntrack to show statistics")
			output, err := mcpInspector.
				MethodCall(getConntrackToolName, map[string]any{
					"node":    nodeName,
					"command": "-S",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains statistics information")
			result := utils.UnmarshalCallToolResult[types.Result](output)
			Expect(result.Data).NotTo(BeEmpty())
			// Conntrack statistics output contains counters
			Expect(result.Data).To(Or(
				ContainSubstring("cpu="),
				ContainSubstring("found="),
				ContainSubstring("invalid="),
			))
		})
	})
})
