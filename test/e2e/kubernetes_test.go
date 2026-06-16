package e2e

import (
	"context"
	"encoding/json"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8se2eframework "k8s.io/kubernetes/test/e2e/framework"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("Kubernetes Tools", func() {
	const (
		resourceGetToolName  = "resource-get"
		resourceListToolName = "resource-list"
	)

	fr := k8se2eframework.NewDefaultFramework("kubernetes-tools")

	// restrictedSecurityContext returns a SecurityContext that complies with
	// OpenShift's "restricted" PodSecurity policy
	restrictedSecurityContext := func() *corev1.SecurityContext {
		allowPrivilegeEscalation := false
		runAsNonRoot := true
		return &corev1.SecurityContext{
			AllowPrivilegeEscalation: &allowPrivilegeEscalation,
			RunAsNonRoot:             &runAsNonRoot,
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
		}
	}

	Context("Get Resource", func() {
		// Replace Secret tests with ConfigMap
		It("should get a configmap from a namespace", func() {
			By("Creating a configmap")
			cmName := "test-configmap"
			err := kubeClient.Create(context.Background(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: fr.Namespace.Name,
				},
				Data: map[string]string{
					"key1": "value1",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Getting the configmap")
			output, err := mcpInspector.
				MethodCall(resourceGetToolName, map[string]any{
					"group":     "",
					"version":   "v1",
					"kind":      "ConfigMap",
					"namespace": fr.Namespace.Name,
					"name":      cmName,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			getResult := utils.UnmarshalCallToolResult[types.GetResourceResult](output)
			Expect(getResult.Resource.Name).To(Equal(cmName))
			Expect(getResult.Resource.Namespace).To(Equal(fr.Namespace.Name))
		})

		It("should get a configmap with JSON output type", func() {
			By("Creating a configmap with data")
			cmName := "test-configmap-json"
			data := map[string]string{
				"app.conf":  "debug=true",
				"db.config": "host=localhost",
			}
			err := kubeClient.Create(context.Background(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: fr.Namespace.Name,
				},
				Data: data,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Getting the configmap with JSON output")
			output, err := mcpInspector.
				MethodCall(resourceGetToolName, map[string]any{
					"group":       "",
					"version":     "v1",
					"kind":        "ConfigMap",
					"namespace":   fr.Namespace.Name,
					"name":        cmName,
					"output_type": string(types.JSONOutputType),
				}).Execute()
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the configmap data")
			getResult := utils.UnmarshalCallToolResult[types.GetResourceResult](output)
			var fetchedCM corev1.ConfigMap
			err = json.Unmarshal([]byte(getResult.Resource.Data), &fetchedCM)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedCM.Data).To(Equal(data))
		})

		It("should get a pod from a namespace", func() {
			By("Creating a pod")
			podName := "test-pod"
			err := kubeClient.Create(context.Background(), &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: fr.Namespace.Name,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "nginx",
							Image:           "nginx:alpine",
							SecurityContext: restrictedSecurityContext(),
						},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Getting the pod")
			output, err := mcpInspector.
				MethodCall(resourceGetToolName, map[string]any{
					"group":     "",
					"version":   "v1",
					"kind":      "Pod",
					"namespace": fr.Namespace.Name,
					"name":      podName,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())

			By("Verifying pod information")
			getResult := utils.UnmarshalCallToolResult[types.GetResourceResult](output)
			Expect(getResult.Resource.Name).To(Equal(podName))
		})

		It("should get a service from a namespace", func() {
			By("Creating a service")
			svcName := "test-service"
			err := kubeClient.Create(context.Background(), &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      svcName,
					Namespace: fr.Namespace.Name,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{"app": "test"},
					Ports: []corev1.ServicePort{
						{Port: 80, Protocol: corev1.ProtocolTCP},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Getting the service")
			output, err := mcpInspector.
				MethodCall(resourceGetToolName, map[string]any{
					"group":     "",
					"version":   "v1",
					"kind":      "Service",
					"namespace": fr.Namespace.Name,
					"name":      svcName,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())

			By("Verifying service information")
			getResult := utils.UnmarshalCallToolResult[types.GetResourceResult](output)
			Expect(getResult.Resource.Name).To(Equal(svcName))
		})

		It("should get a deployment with YAML output type", func() {
			By("Creating a deployment")
			deployName := "test-deployment"
			replicas := int32(1)
			err := kubeClient.Create(context.Background(), &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deployName,
					Namespace: fr.Namespace.Name,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:            "nginx",
									Image:           "nginx:alpine",
									SecurityContext: restrictedSecurityContext(),
								},
							},
						},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Getting the deployment with YAML output")
			output, err := mcpInspector.
				MethodCall(resourceGetToolName, map[string]any{
					"group":       "apps",
					"version":     "v1",
					"kind":        "Deployment",
					"namespace":   fr.Namespace.Name,
					"name":        deployName,
					"output_type": string(types.YAMLOutputType),
				}).Execute()
			Expect(err).NotTo(HaveOccurred())

			By("Verifying YAML output format")
			getResult := utils.UnmarshalCallToolResult[types.GetResourceResult](output)
			Expect(getResult.Resource.Data).To(ContainSubstring("apiVersion: apps/v1"))
			Expect(getResult.Resource.Data).To(ContainSubstring("kind: Deployment"))
		})
	})

	Context("List Resources", func() {
		It("should list configmaps from a namespace", func() {
			By("Creating 2 configmaps")
			cm1 := "test-cm-1"
			cm2 := "test-cm-2"
			err := kubeClient.Create(context.Background(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cm1,
					Namespace: fr.Namespace.Name,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			err = kubeClient.Create(context.Background(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cm2,
					Namespace: fr.Namespace.Name,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Listing configmaps")
			output, err := mcpInspector.
				MethodCall(resourceListToolName, map[string]any{
					"group":     "",
					"version":   "v1",
					"kind":      "ConfigMap",
					"namespace": fr.Namespace.Name,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())

			By("Verifying configmaps are listed")
			listResult := utils.UnmarshalCallToolResult[types.ListResourcesResult](output)
			cmNames := make([]string, len(listResult.Resources))
			for i, resource := range listResult.Resources {
				cmNames[i] = resource.Name
			}
			Expect(cmNames).To(ContainElement(cm1))
			Expect(cmNames).To(ContainElement(cm2))
		})

		It("should list resources with label selector", func() {
			By("Creating configmaps with labels")
			cm1 := "labeled-cm-1"
			cm2 := "labeled-cm-2"
			cm3 := "unlabeled-cm"

			// ConfigMaps with label
			for _, name := range []string{cm1, cm2} {
				err := kubeClient.Create(context.Background(), &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: fr.Namespace.Name,
						Labels:    map[string]string{"env": "test"},
					},
				})
				Expect(err).NotTo(HaveOccurred())
			}

			// ConfigMap without label
			err := kubeClient.Create(context.Background(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cm3,
					Namespace: fr.Namespace.Name,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Listing configmaps with label selector")
			output, err := mcpInspector.
				MethodCall(resourceListToolName, map[string]any{
					"group":          "",
					"version":        "v1",
					"kind":           "ConfigMap",
					"namespace":      fr.Namespace.Name,
					"label_selector": "env=test",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())

			By("Verifying only labeled configmaps are returned")
			listResult := utils.UnmarshalCallToolResult[types.ListResourcesResult](output)
			cmNames := make([]string, len(listResult.Resources))
			for i, resource := range listResult.Resources {
				cmNames[i] = resource.Name
			}
			Expect(cmNames).To(ContainElement(cm1))
			Expect(cmNames).To(ContainElement(cm2))
			Expect(cmNames).NotTo(ContainElement(cm3))
		})

		It("should list pods from a namespace", func() {
			By("Creating test pods")
			pod1 := "test-pod-1"
			pod2 := "test-pod-2"
			for _, name := range []string{pod1, pod2} {
				err := kubeClient.Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: fr.Namespace.Name,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:            "nginx",
								Image:           "nginx:alpine",
								SecurityContext: restrictedSecurityContext(),
							},
						},
					},
				})
				Expect(err).NotTo(HaveOccurred())
			}

			By("Listing pods (resource-list queries API objects regardless of phase)")
			// Note: resource-list reads from Kubernetes API and doesn't require pods to be Running.
			// The tool lists API objects in any phase (Pending/Running/Failed/etc), so no wait needed.
			output, err := mcpInspector.
				MethodCall(resourceListToolName, map[string]any{
					"group":     "",
					"version":   "v1",
					"kind":      "Pod",
					"namespace": fr.Namespace.Name,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())

			By("Verifying pods are listed")
			listResult := utils.UnmarshalCallToolResult[types.ListResourcesResult](output)
			podNames := make([]string, len(listResult.Resources))
			for i, resource := range listResult.Resources {
				podNames[i] = resource.Name
			}
			Expect(podNames).To(ContainElement(pod1))
			Expect(podNames).To(ContainElement(pod2))
		})
	})
})
