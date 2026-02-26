package e2e

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sostypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/sosreport/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

// getTestdataSosreportPath computes the path to the testdata sosreport directory
func getTestdataSosreportPath() string {
	return utils.GetTestdataPath("../../pkg/sosreport/mcp/testdata/sosreport")
}

var _ = Describe("[offline] Sosreport Tools", func() {
	testdataSosreportPath := getTestdataSosreportPath()
	const (
		sosListPluginsToolName    = "sos-list-plugins"
		sosListCommandsToolName   = "sos-list-commands"
		sosSearchCommandsToolName = "sos-search-commands"
		sosGetCommandToolName     = "sos-get-command"
		sosSearchPodLogsToolName  = "sos-search-pod-logs"
	)

	DescribeTable("List Plugins",
		func(sosreportPath string, shouldFail bool) {
			By("Calling sos-list-plugins")
			output, err := mcpInspector.
				MethodCall(sosListPluginsToolName, map[string]string{
					"sosreport_path": sosreportPath,
				}).Execute()

			Expect(err).NotTo(HaveOccurred())
			if shouldFail {
				Expect(string(output)).To(ContainSubstring(`"isError": true`))
			} else {
				Expect(output).NotTo(BeEmpty())

				By("Checking the result")
				result := utils.UnmarshalCallToolResult[sostypes.ListPluginsResult](output)
				Expect(result.Plugins).To(HaveLen(3))
				Expect(result.TotalCommands).To(Equal(3))

				// Verify plugin names and command counts
				pluginMap := make(map[string]int)
				for _, plugin := range result.Plugins {
					pluginMap[plugin.Name] = plugin.CommandCount
				}
				Expect(pluginMap["openvswitch"]).To(Equal(2))
				Expect(pluginMap["networking"]).To(Equal(1))
				Expect(pluginMap["container_log"]).To(Equal(0))
			}
		},
		Entry("should list all enabled plugins with command counts", testdataSosreportPath, false),
		Entry("should fail with non-existent sosreport path", "/path/that/does/not/exist", true),
	)

	DescribeTable("List Commands",
		func(plugin string, expectedCount int, expectedCommands map[string]string, shouldFail bool) {
			By("Calling sos-list-commands for " + plugin)
			output, err := mcpInspector.
				MethodCall(sosListCommandsToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"plugin":         plugin,
				}).Execute()

			Expect(err).NotTo(HaveOccurred())
			if shouldFail {
				Expect(string(output)).To(ContainSubstring(`"isError": true`))
			} else {
				Expect(output).NotTo(BeEmpty())

				By("Checking the result")
				result := utils.UnmarshalCallToolResult[sostypes.ListCommandsResult](output)
				Expect(result.Plugin).To(Equal(plugin))
				Expect(result.CommandCount).To(Equal(expectedCount))
				Expect(result.Commands).To(HaveLen(expectedCount))

				// Verify command details
				commandMap := make(map[string]string)
				for _, cmd := range result.Commands {
					commandMap[cmd.Exec] = cmd.Filepath
				}
				for exec, filepath := range expectedCommands {
					Expect(commandMap[exec]).To(Equal(filepath))
				}
			}
		},
		Entry("should list all commands for openvswitch plugin", "openvswitch", 2, map[string]string{
			"ovs-vsctl -t 5 show":         "sos_commands/openvswitch/ovs-vsctl_-t_5_show",
			"ovs-ofctl dump-flows br-int": "sos_commands/openvswitch/ovs-ofctl_dump-flows_br-int",
		}, false),
		Entry("should list all commands for networking plugin", "networking", 1, map[string]string{
			"ip addr show": "sos_commands/networking/ip_addr_show",
		}, false),
		Entry("should fail with non-existent plugin", "non_existent_plugin", 0, nil, true),
	)

	DescribeTable("Search Commands",
		func(pattern string, expectedTotal int, expectedPlugin string, expectedExecPattern string) {
			By("Calling sos-search-commands with pattern '" + pattern + "'")
			output, err := mcpInspector.
				MethodCall(sosSearchCommandsToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"pattern":        pattern,
				}).Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.SearchCommandsResult](output)
			Expect(result.Total).To(Equal(expectedTotal))
			Expect(result.Matches).To(HaveLen(expectedTotal))

			// Verify all matches are from expected plugin and contain expected pattern
			if expectedTotal > 0 {
				for _, match := range result.Matches {
					Expect(match.Plugin).To(Equal(expectedPlugin))
					Expect(match.Exec).To(ContainSubstring(expectedExecPattern))
				}
			}
		},
		Entry("should find commands matching ovs pattern", "ovs", 2, "openvswitch", "ovs"),
		Entry("should find commands matching ip pattern", "ip", 1, "networking", "ip"),
		Entry("should return empty when pattern matches nothing", "xyz", 0, "", ""),
	)

	DescribeTable("Get Command",
		func(filepath string, expectedSubstrings []string, shouldFail bool) {
			By("Calling sos-get-command")
			output, err := mcpInspector.
				MethodCall(sosGetCommandToolName, map[string]string{
					"sosreport_path": testdataSosreportPath,
					"filepath":       filepath,
				}).Execute()

			Expect(err).NotTo(HaveOccurred())
			if shouldFail {
				Expect(string(output)).To(ContainSubstring(`"isError": true`))
			} else {
				Expect(output).NotTo(BeEmpty())

				By("Checking the result")
				result := utils.UnmarshalCallToolResult[sostypes.GetCommandResult](output)

				// Verify output contains expected content
				for _, expected := range expectedSubstrings {
					Expect(result.Output).To(ContainSubstring(expected))
				}
			}
		},
		Entry("should get ovs-vsctl show command output",
			"sos_commands/openvswitch/ovs-vsctl_-t_5_show",
			[]string{
				"Bridge br-int",
				"ovn-k8s-mp0",
			},
			false,
		),
		Entry("should get full ip addr show output",
			"sos_commands/networking/ip_addr_show",
			[]string{
				"lo",
				"eth0",
				"127.0.0.1",
				"192.168.1.100",
			},
			false,
		),
		Entry("should fail with non-existent filepath",
			"sos_commands/invalid/nonexistent_command",
			nil,
			true,
		),
	)

	It("Get Command should filter command output by pattern", func() {
		By("Calling sos-get-command with pattern filter")
		output, err := mcpInspector.
			MethodCall(sosGetCommandToolName, map[string]string{
				"sosreport_path": testdataSosreportPath,
				"filepath":       "sos_commands/networking/ip_addr_show",
				"pattern":        "eth0",
			}).Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(output).NotTo(BeEmpty())

		By("Checking the result")
		result := utils.UnmarshalCallToolResult[sostypes.GetCommandResult](output)

		// Should only contain lines matching the pattern
		Expect(result.Output).To(ContainSubstring("eth0"))
		// Verify all non-empty lines contain eth0
		iter := strings.SplitSeq(strings.TrimSpace(result.Output), "\n")
		for line := range iter {
			if line != "" {
				Expect(line).To(ContainSubstring("eth0"))
			}
		}
	})

	DescribeTable("Search Pod Logs",
		func(pattern string, podFilter string, expectedSubstrings []string) {
			params := map[string]string{
				"sosreport_path": testdataSosreportPath,
				"pattern":        pattern,
			}
			if podFilter != "" {
				params["pod_filter"] = podFilter
			}

			By("Calling sos-search-pod-logs")
			output, err := mcpInspector.
				MethodCall(sosSearchPodLogsToolName, params).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			result := utils.UnmarshalCallToolResult[sostypes.SearchPodLogsResult](output)

			// Verify all expected substrings are present
			for _, expected := range expectedSubstrings {
				Expect(result.Output).To(ContainSubstring(expected))
			}
		},
		Entry("should search for ERROR pattern in pod logs",
			"ERROR",
			"",
			[]string{
				"ERROR",
				"Failed to connect to ovn-controller",
				"ovnkube-node",
			},
		),
		Entry("should search pod logs with pod filter",
			"ovnkube",
			"ovnkube-node",
			[]string{
				"ovnkube-node",
				"Starting ovnkube-node",
			},
		),
		Entry("should search for successful connection in pod logs",
			"Successfully connected",
			"",
			[]string{
				"Successfully connected to ovn-controller",
			},
		),
	)
})
