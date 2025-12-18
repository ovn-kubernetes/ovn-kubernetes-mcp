package e2e

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/ovn/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("OVN Tools", func() {
	const (
		ovnShowToolName = "ovn-show"
	)

	// Common OVN-Kubernetes namespaces to search for ovnkube-node pods
	var ovnNamespaces = []string{
		"openshift-ovn-kubernetes",
		"ovn-kubernetes",
	}

	// findOVNKubeNodePod finds a running ovnkube-node pod
	findOVNKubeNodePod := func() (namespace, name string) {
		for _, ns := range ovnNamespaces {
			podList := &corev1.PodList{}
			err := kubeClient.List(context.Background(), podList,
				client.InNamespace(ns),
				client.MatchingLabels{"app": "ovnkube-node"},
			)
			if err != nil {
				continue
			}
			for _, pod := range podList.Items {
				if pod.Status.Phase == corev1.PodRunning {
					return ns, pod.Name
				}
			}
		}
		return "", ""
	}

	Context("OVN Show", func() {
		It("should show OVN Northbound database configuration", func() {
			By("Finding an ovnkube-node pod")
			namespace, podName := findOVNKubeNodePod()
			Expect(namespace).NotTo(BeEmpty(), "No running ovnkube-node pod found")
			Expect(podName).NotTo(BeEmpty(), "No running ovnkube-node pod found")

			By("Running ovn-show for NBDB")
			output, err := mcpInspector.
				MethodCall(ovnShowToolName, map[string]string{
					"namespace": namespace,
					"name":      podName,
					"database":  "nbdb",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			showResult := utils.UnmarshalCallToolResult[types.ShowResult](output)
			Expect(showResult.Database).To(Equal(types.NorthboundDB))
			Expect(showResult.Output).NotTo(BeEmpty())
			// NBDB show output typically contains "switch" for logical switches
			Expect(showResult.Output).To(Or(
				ContainSubstring("switch"),
				ContainSubstring("router"),
			))
		})

		It("should show OVN Southbound database configuration", func() {
			By("Finding an ovnkube-node pod")
			namespace, podName := findOVNKubeNodePod()
			Expect(namespace).NotTo(BeEmpty(), "No running ovnkube-node pod found")
			Expect(podName).NotTo(BeEmpty(), "No running ovnkube-node pod found")

			By("Running ovn-show for SBDB")
			output, err := mcpInspector.
				MethodCall(ovnShowToolName, map[string]string{
					"namespace": namespace,
					"name":      podName,
					"database":  "sbdb",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			showResult := utils.UnmarshalCallToolResult[types.ShowResult](output)
			Expect(showResult.Database).To(Equal(types.SouthboundDB))
			Expect(showResult.Output).NotTo(BeEmpty())
			// SBDB show output typically contains "Chassis" information
			Expect(showResult.Output).To(ContainSubstring("Chassis"))
		})
	})
})
