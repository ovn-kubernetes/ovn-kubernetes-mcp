// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"maps"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mcpKernel "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kernel/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kernel/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("Kernel Tools", Ordered, func() {
	const (
		getIPToolName        = "get-ip"
		getIPTablesToolName  = "get-iptables"
		getNFTToolName       = "get-nft"
		getConntrackToolName = "get-conntrack"
	)

	var nodeName string

	BeforeAll(func() {
		var err error
		nodeName, err = utils.FindReadyNode(kubeClient)
		Expect(err).NotTo(HaveOccurred())
	})

	withNode := func(args map[string]any) map[string]any {
		toolArgs := make(map[string]any, len(args)+1)
		maps.Copy(toolArgs, args)
		toolArgs["node"] = nodeName
		return toolArgs
	}

	Context("get-ip", func() {
		It("should retrieve IP routing information from a node", func() {
			By("Running get-ip to show routes")
			output, err := mcpInspector.
				MethodCall(getIPToolName, withNode(map[string]any{
					"command": "route show",
				})).Execute()
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
				MethodCall(getIPToolName, withNode(map[string]any{
					"command": "link show",
				})).Execute()
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
				MethodCall(getIPTablesToolName, withNode(map[string]any{
					"table":   "filter",
					"command": "-L",
				})).Execute()
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
				MethodCall(getIPTablesToolName, withNode(map[string]any{
					"table":   "nat",
					"command": "-L",
				})).Execute()
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
				MethodCall(getNFTToolName, withNode(map[string]any{
					"command": "list tables",
				})).Execute()
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
				utils.ExpectToolError(mcpInspector, toolName, withNode(toolArgs), wantError)
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
				utils.ExpectToolError(mcpInspector, toolName, withNode(toolArgs), "invalid use of metacharacters in parameter")
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
		It("should retrieve connection tracking list from a node", func() {
			By("Running get-conntrack to list connections")
			output, err := mcpInspector.
				MethodCall(getConntrackToolName, withNode(map[string]any{
					"command": "-L",
				})).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result contains connection list")
			result := utils.UnmarshalCallToolResult[types.Result](output)
			Expect(result.Data).NotTo(BeEmpty())
			// List output contains connection information from conntrack CLI or /proc/net/nf_conntrack.
			// At least one marker must be present.
			conntrackListMarkers := []string{
				"src=", "dst=", "sport=", "dport=", "proto=", "state=",
				"ESTABLISHED", "TIME_WAIT", "CLOSE",
			}
			Expect(result.Data).To(Satisfy(func(data string) bool {
				for _, marker := range conntrackListMarkers {
					if strings.Contains(data, marker) {
						return true
					}
				}
				return false
			}))

			// If the summary is present, check if it matches the expected format
			// The summary format is as follows:
			// -- conntrack summary --
			// conntrack v<version> (conntrack-tools): <count> flow entries have been shown.
			if strings.Contains(result.Data, "-- conntrack summary --") {
				By("Checking the summary matches the expected format")
				lines := strings.Split(strings.TrimSpace(result.Data), "\n")
				summary := lines[len(lines)-1]
				Expect(summary).To(MatchRegexp(mcpKernel.ConntrackSummaryPattern.String()))
			}
		})

		It("should retrieve connection tracking count from a node", func() {
			By("Running get-conntrack to count connections")
			output, err := mcpInspector.
				MethodCall(getConntrackToolName, withNode(map[string]any{
					"command": "-C",
				})).Execute()
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
				MethodCall(getConntrackToolName, withNode(map[string]any{
					"command": "-S",
				})).Execute()
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
