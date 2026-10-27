// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"maps"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"
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
		DescribeTable("should preserve diagnostic global options",
			func(options string, expected string) {
				output, err := mcpInspector.MethodCall(getIPToolName, withNode(map[string]any{
					"options": options, "command": "link show", "filter_parameters": "dev lo",
				})).Execute()
				Expect(err).NotTo(HaveOccurred())
				result := utils.UnmarshalCallToolResult[types.Result](output)
				Expect(result.Data).To(ContainSubstring(expected))
			},
			Entry("short detail and statistics options", "-s -d -o", "lo"),
			Entry("long JSON formatting options", "--json --pretty --details", `"ifname": "lo"`),
			Entry("brief output", "-br", "lo"),
			Entry("family and receive buffer options", "-family link -rcvbuf 1048576 -color=never", "lo"),
		)

		DescribeTable("should reject unsafe or invalid global options",
			func(options, wantError string) {
				// A missed validation must not run injected commands on a node.
				args := withNode(map[string]any{
					"namespace": "e2e-kernel-validation-" + uuid.NewString(),
					"command":   "link show",
					"options":   options,
				})
				utils.ExpectToolError(mcpInspector, getIPToolName, args, wantError)
			},
			Entry("batch mode", "-batch /tmp/ovnk-e2e-batch", "invalid ip option"),
			Entry("batch mode with two hyphens", "--batch /tmp/ovnk-e2e-batch", "invalid ip option"),
			Entry("short batch option", "-b /tmp/ovnk-e2e-batch", "invalid ip option"),
			Entry("abbreviated batch option", "-ba /tmp/ovnk-e2e-batch", "invalid ip option"),
			Entry("positional command", "link set lo down", "invalid ip option"),
			Entry("command after option separator", "-- link set lo down", "invalid ip option"),
			Entry("missing namespace", "-n", "missing or invalid value for ip option"),
			Entry("missing family", "-family", "missing or invalid value for ip option"),
			Entry("missing receive buffer size", "-rcvbuf", "missing or invalid value for ip option"),
			Entry("option as namespace value", "-n -batch /tmp/ovnk-e2e-batch", "missing or invalid value for ip option"),
			Entry("unknown family", "-family bogus", "invalid ip address family"),
			Entry("nonnumeric receive buffer size", "-rcvbuf nope", "invalid ip receive buffer size"),
			Entry("overflowing receive buffer size", "-rcvbuf 4294967296", "invalid ip receive buffer size"),
			Entry("invalid color", "-color=bogus", "invalid ip option"),
			Entry("value on a flag", "-json=true", "invalid ip option"),
			Entry("bundled short flags", "-sd", "invalid ip option"),
		)

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
		It("should limit rule listing to the requested chain", func() {
			output, err := mcpInspector.MethodCall(getIPTablesToolName, withNode(map[string]any{
				"table": "filter", "command": "-S", "filter_parameters": "INPUT", "tail": -1,
			})).Execute()
			Expect(err).NotTo(HaveOccurred())
			result := utils.UnmarshalCallToolResult[types.Result](output)
			Expect(result.Data).NotTo(BeEmpty())
			for line := range strings.SplitSeq(result.Data, "\n") {
				fields := strings.Fields(line)
				Expect(len(fields)).To(BeNumerically(">=", 2), "unexpected rule listing line: %q", line)
				Expect(fields[0]).To(BeElementOf("-P", "-A"))
				Expect(fields[1]).To(Equal("INPUT"))
			}
		})

		DescribeTable("should preserve native listing syntax",
			func(filters string) {
				output, err := mcpInspector.MethodCall(getIPTablesToolName, withNode(map[string]any{
					"table": "filter", "command": "-L", "filter_parameters": filters,
				})).Execute()
				Expect(err).NotTo(HaveOccurred())
				result := utils.UnmarshalCallToolResult[types.Result](output)
				Expect(result.Data).To(ContainSubstring("Chain INPUT"))
			},
			Entry("combined short flags", "INPUT -nv4"),
			Entry("IPv6 listing", "INPUT -nv6"),
			Entry("long listing options", "INPUT --numeric --exact --line-numbers"),
			Entry("wait with separate value", "INPUT -w 1 -n4"),
			Entry("wait with attached value", "INPUT --wait=1 -n4"),
		)

		DescribeTable("should reject unsafe or invalid listing filters",
			func(command, filters, wantError string) {
				// A missed validation must not change rules or counters on a node.
				args := withNode(map[string]any{
					"namespace":         "e2e-kernel-validation-" + uuid.NewString(),
					"table":             "filter",
					"command":           command,
					"filter_parameters": filters,
				})
				utils.ExpectToolError(mcpInspector, getIPTablesToolName, args, wantError)
			},
			Entry("counter clearing with default command", "", "-Z", "invalid iptables filter option"),
			Entry("long counter clearing option", "-L", "--zero", "invalid iptables filter option"),
			Entry("counter clearing in a short bundle", "-L", "-nvZ", "invalid iptables filter option"),
			Entry("module helper", "-L", "--modprobe=/bin/false", "invalid iptables filter option"),
			Entry("table override", "-L", "--table=nat", "invalid iptables filter option"),
			Entry("alternate operation", "-L", "-A INPUT", "invalid iptables filter option"),
			Entry("unsupported protocol selector", "-L", "-p tcp", "invalid iptables filter option"),
			Entry("abbreviated long option", "-L", "--num", "invalid iptables filter option"),
			Entry("value on a flag", "-L", "--numeric=true", "invalid iptables filter option"),
			Entry("empty wait value", "-L", "--wait=", "invalid iptables filter option"),
			Entry("nonnumeric wait value", "-L", "-w nope", "invalid value for iptables filter option"),
			Entry("overflowing wait value", "-L", "-w 2147483648", "invalid value for iptables filter option"),
			Entry("invalid chain", "-L", "!INPUT", "invalid iptables chain"),
			Entry("zero rule number", "-L", "INPUT 0", "invalid iptables rule number"),
			Entry("overflowing rule number", "-L", "INPUT 2147483648", "invalid iptables rule number"),
			Entry("chain after options", "-L", "-n INPUT", "unexpected iptables filter argument"),
			Entry("lone hyphen", "-L", "-", "invalid iptables filter option"),
			Entry("option separator", "-L", "--", "invalid iptables filter option"),
			Entry("counter clearing with rule listing", "-S", "-Z", "invalid iptables filter option"),
			Entry("numeric output with rule listing", "-S", "-n", "is not supported with -S"),
			Entry("line numbers with long rule listing", "--list-rules", "INPUT --line-numbers", "is not supported with --list-rules"),
		)

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
		It("should reject filters for connection counts", func() {
			utils.ExpectToolError(mcpInspector, getConntrackToolName, withNode(map[string]any{
				"command": "-C", "filter_parameters": "-p udp",
			}), "filter_parameters is only supported with -L/--dump")
		})

		DescribeTable("should preserve read-only dump filters",
			func(filters string) {
				output, err := mcpInspector.MethodCall(getConntrackToolName, withNode(map[string]any{
					"command": "-L", "filter_parameters": filters,
				})).Execute()
				Expect(err).NotTo(HaveOccurred())
				result := utils.UnmarshalCallToolResult[types.Result](output)
				Expect(result.Data).To(ContainSubstring("flow entries have been shown."))
			},
			Entry("protocol and port", "--proto tcp --dport 443"),
			Entry("attached values", "-s192.0.2.1 -ptcp --dport=443"),
			Entry("table, family and output", "conntrack --family ipv4 -o extended,id"),
			Entry("NAT filter", "--src-nat"),
		)

		DescribeTable("should reject unsafe or invalid dump filters",
			func(command, filters, wantError string) {
				// A missed validation must not change connection tracking on a node.
				args := withNode(map[string]any{
					"namespace":         "e2e-kernel-validation-" + uuid.NewString(),
					"command":           command,
					"filter_parameters": filters,
				})
				utils.ExpectToolError(mcpInspector, getConntrackToolName, args, wantError)
			},
			Entry("counter clearing with default command", "", "-z", "invalid conntrack filter option"),
			Entry("long counter clearing option", "--dump", "--zero", "invalid conntrack filter option"),
			Entry("counter clearing after optional NAT filter", "-L", "--src-nat --zero", "invalid conntrack filter option"),
			Entry("counter clearing in a short bundle", "-L", "-zn", "invalid conntrack filter option"),
			Entry("alternate operation", "-L", "--delete", "invalid conntrack filter option"),
			Entry("abbreviated filter", "-L", "--prot tcp", "invalid conntrack filter option"),
			Entry("missing protocol", "-L", "-p", "missing value for conntrack filter option"),
			Entry("empty address", "-L", "--src=", "invalid value for conntrack filter option"),
			Entry("option as attached address", "-L", "--src=-z", "invalid value for conntrack filter option"),
			Entry("option as short protocol value", "-L", "-p-z", "invalid value for conntrack filter option"),
			Entry("option as separate port value", "-L", "--sport --zero", "missing value for conntrack filter option"),
			Entry("empty optional NAT value", "-L", "--any-nat=", "invalid value for conntrack filter option"),
			Entry("negated optional NAT value", "-L", "--src-nat !192.0.2.1", "invalid value for conntrack filter option"),
			Entry("unexpected positional argument", "-L", "-p tcp extra", "invalid conntrack filter option"),
			Entry("option separator", "-L", "--", "invalid conntrack filter option"),
			Entry("counter clearing with statistics", "--stats", "--zero", "invalid conntrack filter option"),
		)

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
