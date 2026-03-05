package e2e

import (
	"context"
	"time"

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
		debugImage           = "nicolaka/netshoot:v0.13"
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

	Context("get-ip", func() {
		It("should retrieve IP routing information from a node", func() {
			By("Running get-ip to show routes")
			output, err := mcpInspector.
				MethodCall(getIPToolName, map[string]string{
					"node":    nodeName,
					"image":   debugImage,
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
				MethodCall(getIPToolName, map[string]string{
					"node":    nodeName,
					"image":   debugImage,
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
				MethodCall(getIPTablesToolName, map[string]string{
					"node":    nodeName,
					"image":   debugImage,
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
				MethodCall(getIPTablesToolName, map[string]string{
					"node":    nodeName,
					"image":   debugImage,
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
				MethodCall(getNFTToolName, map[string]string{
					"node":    nodeName,
					"image":   debugImage,
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

	Context("get-conntrack", func() {
		It("should retrieve connection tracking count from a node", func() {
			By("Running get-conntrack to count connections")
			output, err := mcpInspector.
				MethodCall(getConntrackToolName, map[string]string{
					"node":    nodeName,
					"image":   debugImage,
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
				MethodCall(getConntrackToolName, map[string]string{
					"node":    nodeName,
					"image":   debugImage,
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
