// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"maps"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8se2eframework "k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/network-tools/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("Network Tools", Ordered, func() {
	const tcpdumpToolName = "tcpdump"

	fr := k8se2eframework.NewDefaultFramework("network-tools")

	var nodeName string

	BeforeAll(func() {
		var err error
		nodeName, err = utils.FindReadyNode(kubeClient)
		Expect(err).NotTo(HaveOccurred())
	})

	callTcpdump := func(args map[string]any) types.CommandResult {
		toolArgs := make(map[string]any, len(args)+1)
		maps.Copy(toolArgs, args)
		toolArgs["timeout_seconds"] = 30
		var result types.CommandResult
		var lastErr string
		Eventually(func() bool {
			output, err := mcpInspector.
				MethodCall(tcpdumpToolName, toolArgs).Execute()
			if err != nil {
				lastErr = fmt.Sprintf("execute error: %v", err)
				return false
			}
			if len(output) == 0 {
				lastErr = "empty output from MCP server"
				return false
			}
			var callResult mcp.CallToolResult
			if err := callResult.UnmarshalJSON(output); err != nil {
				lastErr = fmt.Sprintf("unmarshal error: %v", err)
				return false
			}
			if callResult.IsError {
				lastErr = fmt.Sprintf("tool returned error: %s", string(output))
				return false
			}
			result = utils.UnmarshalCallToolResult[types.CommandResult](output)
			return result.Output != ""
		}, 120*time.Second, 10*time.Second).Should(BeTrue(),
			fmt.Sprintf("tcpdump never returned successfully; last error: %s", lastErr))
		return result
	}

	createServerPod := func(name string) string {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: fr.Namespace.Name,
			},
			Spec: corev1.PodSpec{
				NodeName: nodeName,
				Containers: []corev1.Container{
					{
						Name:            "server",
						Image:           "busybox:1.36",
						Command:         []string{"/bin/sh", "-c", "mkdir -p /tmp/www && echo ok > /tmp/www/index.html && httpd -f -p 8080 -h /tmp/www"},
						Ports:           []corev1.ContainerPort{{ContainerPort: 8080}},
						SecurityContext: utils.RestrictedSecurityContext(),
					},
				},
			},
		}
		err := kubeClient.Create(context.Background(), pod)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() bool {
			err := kubeClient.Get(context.Background(), client.ObjectKey{
				Namespace: fr.Namespace.Name,
				Name:      name,
			}, pod)
			if err != nil {
				return false
			}
			return pod.Status.Phase == corev1.PodRunning && pod.Status.PodIP != ""
		}, 60*time.Second, 2*time.Second).Should(BeTrue())

		return pod.Status.PodIP
	}

	generateTraffic := func(serverIP string) {
		clientPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "traffic-client",
				Namespace: fr.Namespace.Name,
			},
			Spec: corev1.PodSpec{
				NodeName:      nodeName,
				RestartPolicy: corev1.RestartPolicyNever,
				Containers: []corev1.Container{
					{
						Name:            "client",
						Image:           "busybox:1.36",
						Command:         []string{"/bin/sh", "-c", fmt.Sprintf("while true; do wget -q -T 2 -O /dev/null http://%s/ 2>/dev/null; sleep 1; done", net.JoinHostPort(serverIP, "8080"))},
						SecurityContext: utils.RestrictedSecurityContext(),
					},
				},
			},
		}
		err := kubeClient.Create(context.Background(), clientPod)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() bool {
			err := kubeClient.Get(context.Background(), client.ObjectKey{
				Namespace: fr.Namespace.Name,
				Name:      "traffic-client",
			}, clientPod)
			if err != nil {
				return false
			}
			return clientPod.Status.Phase == corev1.PodRunning
		}, 60*time.Second, 2*time.Second).Should(BeTrue())
	}

	Context("tcpdump", func() {
		var serverIP string

		BeforeEach(func() {
			By("Creating a server pod")
			serverIP = createServerPod("tcpdump-server")

			By("Generating pod-to-pod traffic")
			generateTraffic(serverIP)
		})

		AfterEach(func() {
			for _, name := range []string{"tcpdump-server", "traffic-client"} {
				Expect(client.IgnoreNotFound(kubeClient.Delete(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: fr.Namespace.Name},
				}))).NotTo(HaveOccurred())
			}
			for _, name := range []string{"tcpdump-server", "traffic-client"} {
				podName := name
				Eventually(func() bool {
					err := kubeClient.Get(context.Background(), client.ObjectKey{
						Namespace: fr.Namespace.Name,
						Name:      podName,
					}, &corev1.Pod{})
					return client.IgnoreNotFound(err) == nil && err != nil
				}, 60*time.Second, 2*time.Second).Should(BeTrue(),
					"timed out waiting for pod %s to be deleted", podName)
			}
		})

		It("should capture pod-to-pod traffic", func() {
			By("Running tcpdump to capture network traffic")
			result := callTcpdump(map[string]any{
				"target_type":  "node",
				"name":         nodeName,
				"packet_count": 5,
				"interface":    "any",
				"bpf_filter":   fmt.Sprintf("host %s and port 8080", serverIP),
			})

			By("Verifying tcpdump captured traffic")
			Expect(result.Output).NotTo(BeEmpty())
			Expect(result.Output).To(ContainSubstring(serverIP))
		})

		It("should capture traffic with BPF filter", func() {
			By("Running tcpdump with TCP filter")
			result := callTcpdump(map[string]any{
				"target_type":  "node",
				"name":         nodeName,
				"packet_count": 5,
				"interface":    "any",
				"bpf_filter":   fmt.Sprintf("tcp and host %s and port 8080", serverIP),
			})

			By("Verifying tcpdump BPF filter output")
			Expect(result.Output).NotTo(BeEmpty())
			Expect(result.Output).To(ContainSubstring(serverIP))
			Expect(strings.ToLower(result.Output)).To(ContainSubstring("tcp"))
		})

		It("should respect packet_count limit", func() {
			By("Running tcpdump with packet_count=2")
			result := callTcpdump(map[string]any{
				"target_type":  "node",
				"name":         nodeName,
				"packet_count": 2,
				"interface":    "any",
				"bpf_filter":   fmt.Sprintf("host %s and port 8080", serverIP),
				"snaplen":      128,
			})

			By("Verifying packet_count=2 limits the number of captured packets")
			Expect(result.Output).NotTo(BeEmpty())
			Expect(result.Output).To(ContainSubstring(serverIP))
			packetLineRe := regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d+ IP`)
			var packetLines []string
			for _, line := range strings.Split(strings.TrimSpace(result.Output), "\n") {
				if packetLineRe.MatchString(line) {
					packetLines = append(packetLines, line)
				}
			}
			Expect(len(packetLines)).To(BeNumerically("<=", 2),
				"packet_count=2 should produce at most 2 captured packet lines")
		})
	})

	Context("validation errors", func() {
		DescribeTable("should reject invalid tcpdump inputs",
			func(toolArgs map[string]any, wantError string) {
				utils.ExpectToolError(mcpInspector, tcpdumpToolName, toolArgs, wantError)
			},
			Entry("missing target_type", map[string]any{
				"name": "some-node",
			}, "target_type"),
			Entry("invalid target_type", map[string]any{
				"target_type": "invalid",
				"name":        "some-node",
			}, "target_type"),
			Entry("missing name", map[string]any{
				"target_type": "node",
			}, "name"),
			Entry("pod without namespace", map[string]any{
				"target_type": "pod",
				"name":        "some-pod",
			}, "namespace is required"),
			Entry("invalid interface name", map[string]any{
				"target_type": "node",
				"name":        "some-node",
				"interface":   "eth0; rm -rf /",
			}, "invalid interface name"),
			Entry("packet_count exceeds max", map[string]any{
				"target_type":  "node",
				"name":         "some-node",
				"packet_count": 9999,
			}, "packet_count cannot exceed"),
			Entry("snaplen exceeds max", map[string]any{
				"target_type": "node",
				"name":        "some-node",
				"snaplen":     9999,
			}, "snaplen cannot exceed"),
			Entry("negative packet_count", map[string]any{
				"target_type":  "node",
				"name":         "some-node",
				"packet_count": -1,
			}, "packet_count cannot be negative"),
			Entry("negative snaplen", map[string]any{
				"target_type": "node",
				"name":        "some-node",
				"snaplen":     -1,
			}, "snaplen cannot be negative"),
			Entry("interface name too long", map[string]any{
				"target_type": "node",
				"name":        "some-node",
				"interface":   "abcdefghijklmnop",
			}, "interface name too long"),
			Entry("bpf_filter too long", map[string]any{
				"target_type": "node",
				"name":        "some-node",
				"bpf_filter":  strings.Repeat("a", 1025),
			}, "packet filter too long"),
		)

		DescribeTable("should reject metacharacters in filters",
			func(toolName string, toolArgs map[string]any) {
				utils.ExpectToolError(mcpInspector, toolName, toolArgs, "invalid use of metacharacters")
			},
			Entry("tcpdump bpf_filter", tcpdumpToolName, map[string]any{
				"target_type": "node",
				"name":        "some-node",
				"bpf_filter":  "tcp; rm -rf /",
			}),
		)
	})

})
