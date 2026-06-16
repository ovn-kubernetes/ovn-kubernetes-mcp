package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8se2eframework "k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("Kubernetes Tools", func() {
	const (
		resourceGetToolName  = "resource-get"
		resourceListToolName = "resource-list"
		podLogsToolName      = "pod-logs"
	)

	fr := k8se2eframework.NewDefaultFramework("kubernetes-tools")

	// restrictedSecurityContext returns a SecurityContext that complies with
	// the "restricted" PodSecurity standard
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
			Expect(output).NotTo(BeEmpty())

			By("Verifying the configmap data")
			getResult := utils.UnmarshalCallToolResult[types.GetResourceResult](output)
			var fetchedCM corev1.ConfigMap
			err = json.Unmarshal([]byte(getResult.Resource.Data), &fetchedCM)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedCM.Data).To(Equal(data))
		})

		It("should get a pod with jsonTemplate to extract container image", func() {
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

			By("Getting the pod with jsonTemplate to extract container image")
			output, err := mcpInspector.
				MethodCall(resourceGetToolName, map[string]any{
					"group":         "",
					"version":       "v1",
					"kind":          "Pod",
					"namespace":     fr.Namespace.Name,
					"name":          podName,
					"output_type": "jsonpath={.spec.containers[0].image}",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Verifying extracted container image")
			getResult := utils.UnmarshalCallToolResult[types.GetResourceResult](output)
			Expect(getResult.Resource.Data).To(ContainSubstring("nginx:alpine"))
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
			Expect(output).NotTo(BeEmpty())

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
			Expect(output).NotTo(BeEmpty())

			By("Verifying YAML output contains correct deployment spec")
			getResult := utils.UnmarshalCallToolResult[types.GetResourceResult](output)

			var fetchedDeploy appsv1.Deployment
			err = yaml.Unmarshal([]byte(getResult.Resource.Data), &fetchedDeploy)
			Expect(err).NotTo(HaveOccurred())

			Expect(fetchedDeploy.Name).To(Equal(deployName))
			Expect(fetchedDeploy.Namespace).To(Equal(fr.Namespace.Name))
			Expect(*fetchedDeploy.Spec.Replicas).To(Equal(replicas))
			Expect(fetchedDeploy.Spec.Selector.MatchLabels).To(Equal(map[string]string{"app": "test"}))
			Expect(fetchedDeploy.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(fetchedDeploy.Spec.Template.Spec.Containers[0].Name).To(Equal("nginx"))
			Expect(fetchedDeploy.Spec.Template.Spec.Containers[0].Image).To(Equal("nginx:alpine"))
		})
	})

	Context("List Resources", func() {
		It("should list configmaps from a namespace", func() {
			By("Creating 2 configmaps")
			cm1 := "test-cm-1"
			cm2 := "test-cm-2"
			testLabel := map[string]string{"test": "e2e-list-cm"}
			err := kubeClient.Create(context.Background(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cm1,
					Namespace: fr.Namespace.Name,
					Labels:    testLabel,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			err = kubeClient.Create(context.Background(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cm2,
					Namespace: fr.Namespace.Name,
					Labels:    testLabel,
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
					"label_selector": "test=e2e-list-cm",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Verifying configmaps are listed")
			listResult := utils.UnmarshalCallToolResult[types.ListResourcesResult](output)
			Expect(listResult.Resources).To(HaveLen(2))
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
			Expect(output).NotTo(BeEmpty())

			By("Verifying only labeled configmaps are returned")
			listResult := utils.UnmarshalCallToolResult[types.ListResourcesResult](output)
			Expect(listResult.Resources).To(HaveLen(2))
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
			testLabel := map[string]string{"test": "e2e-list-pods"}
			for _, name := range []string{pod1, pod2} {
				err := kubeClient.Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: fr.Namespace.Name,
						Labels:    testLabel,
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

			By("Listing pods with label selector (resource-list queries API objects regardless of phase)")
			// Note: resource-list reads from Kubernetes API and doesn't require pods to be Running.
			// The tool lists API objects in any phase (Pending/Running/Failed/etc), so no wait needed.
			output, err := mcpInspector.
				MethodCall(resourceListToolName, map[string]any{
					"group":          "",
					"version":        "v1",
					"kind":           "Pod",
					"namespace":      fr.Namespace.Name,
					"label_selector": "test=e2e-list-pods",
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Verifying pods are listed")
			listResult := utils.UnmarshalCallToolResult[types.ListResourcesResult](output)
			Expect(listResult.Resources).To(HaveLen(2))
			podNames := make([]string, len(listResult.Resources))
			for i, resource := range listResult.Resources {
				podNames[i] = resource.Name
			}
			Expect(podNames).To(ContainElement(pod1))
			Expect(podNames).To(ContainElement(pod2))
		})

		It("should list configmaps with JSON output type", func() {
			By("Creating configmaps with data")
			cm1 := "test-cm-json-1"
			cm2 := "test-cm-json-2"
			testLabel := map[string]string{"test": "e2e-json-cm"}

			for i, name := range []string{cm1, cm2} {
				err := kubeClient.Create(context.Background(), &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: fr.Namespace.Name,
						Labels:    testLabel,
					},
					Data: map[string]string{
						"key": fmt.Sprintf("value%d", i+1),
					},
				})
				Expect(err).NotTo(HaveOccurred())
			}

			By("Listing configmaps with JSON output and label selector")
			output, err := mcpInspector.
				MethodCall(resourceListToolName, map[string]any{
					"group":          "",
					"version":        "v1",
					"kind":           "ConfigMap",
					"namespace":      fr.Namespace.Name,
					"label_selector": "test=e2e-json-cm",
					"output_type":    string(types.JSONOutputType),
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Verifying JSON output contains configmap data")
			listResult := utils.UnmarshalCallToolResult[types.ListResourcesResult](output)
			Expect(listResult.Resources).To(HaveLen(2))

			// Each resource should have JSON data - parse and verify
			var foundNames []string
			for _, resource := range listResult.Resources {
				Expect(resource.Data).NotTo(BeEmpty())

				var cm corev1.ConfigMap
				err = json.Unmarshal([]byte(resource.Data), &cm)
				Expect(err).NotTo(HaveOccurred())
				foundNames = append(foundNames, cm.Name)
			}

			// Verify ConfigMap names are present
			Expect(foundNames).To(ContainElement(cm1))
			Expect(foundNames).To(ContainElement(cm2))
		})
	})

	Context("Pod Logs", func() {
		It("should get logs from a running pod", func() {
			By("Creating a pod that generates logs")
			podName := "logging-pod"
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: fr.Namespace.Name,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "logger",
							Image:   "busybox:1.36",
							Command: []string{"/bin/sh", "-c"},
							Args: []string{
								"for i in 1 2 3 4 5 6 7 8 9 10; do echo 'Log line '$i; sleep 1; done; sleep 30",
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: func() *bool { b := false; return &b }(),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
								RunAsNonRoot: func() *bool { b := true; return &b }(),
								RunAsUser:    func() *int64 { i := int64(1000); return &i }(),
								SeccompProfile: &corev1.SeccompProfile{
									Type: corev1.SeccompProfileTypeRuntimeDefault,
								},
							},
						},
					},
				},
			}
			err := kubeClient.Create(context.Background(), pod)
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for pod to be running")
			Eventually(func() bool {
				err := kubeClient.Get(context.Background(), client.ObjectKey{
					Namespace: fr.Namespace.Name,
					Name:      podName,
				}, pod)
				if err != nil {
					return false
				}
				return pod.Status.Phase == corev1.PodRunning
			}, 60*time.Second, 2*time.Second).Should(BeTrue())

			By("Waiting for logs to be generated")
			time.Sleep(3 * time.Second)

			By("Getting pod logs")
			output, err := mcpInspector.
				MethodCall(podLogsToolName, map[string]any{
					"namespace": fr.Namespace.Name,
					"name":      podName,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Verifying logs are returned")
			logsResult := utils.UnmarshalCallToolResult[types.GetPodLogsResult](output)
			Expect(logsResult.Logs).NotTo(BeEmpty())
		})

		It("should get logs with tail parameter", func() {
			By("Creating a pod that generates logs")
			podName := "tail-logs-pod"
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: fr.Namespace.Name,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "logger",
							Image:   "busybox:1.36",
							Command: []string{"/bin/sh", "-c"},
							Args: []string{
								"for i in 1 2 3 4 5 6 7 8 9 10; do echo 'Log line '$i; sleep 1; done; sleep 30",
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: func() *bool { b := false; return &b }(),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
								RunAsNonRoot: func() *bool { b := true; return &b }(),
								RunAsUser:    func() *int64 { i := int64(1000); return &i }(),
								SeccompProfile: &corev1.SeccompProfile{
									Type: corev1.SeccompProfileTypeRuntimeDefault,
								},
							},
						},
					},
				},
			}
			err := kubeClient.Create(context.Background(), pod)
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for pod to be running")
			Eventually(func() bool {
				err := kubeClient.Get(context.Background(), client.ObjectKey{
					Namespace: fr.Namespace.Name,
					Name:      podName,
				}, pod)
				if err != nil {
					return false
				}
				return pod.Status.Phase == corev1.PodRunning
			}, 60*time.Second, 2*time.Second).Should(BeTrue())

			By("Waiting for logs to be generated")
			time.Sleep(6 * time.Second)

			By("Getting pod logs with tail=5")
			output, err := mcpInspector.
				MethodCall(podLogsToolName, map[string]any{
					"namespace": fr.Namespace.Name,
					"name":      podName,
					"tail":      5,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Verifying logs are returned with tail limit")
			logsResult := utils.UnmarshalCallToolResult[types.GetPodLogsResult](output)
			Expect(logsResult.Logs).NotTo(BeEmpty())
			Expect(len(logsResult.Logs)).To(BeNumerically("<=", 5))
		})
	})
})
