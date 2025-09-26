package e2e

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8se2eframework "k8s.io/kubernetes/test/e2e/framework"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/test/e2e/utils"
)

var _ = Describe("Kubernetes Tools", func() {
	const (
		resourceGetToolName  = "resource-get"
		resourceListToolName = "resource-list"
	)

	fr := k8se2eframework.NewDefaultFramework("kubernetes-tools")

	Context("Get Resource", func() {
		It("should get a secret from a namespace", func() {
			By("Creating a secret")
			secretName := "test-secret"
			err := kubeClient.Create(context.Background(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: fr.Namespace.Name,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Getting the secret")
			output, err := mcpInspector.
				MethodCall(resourceGetToolName, map[string]string{
					"group":     "",
					"version":   "v1",
					"kind":      "Secret",
					"namespace": fr.Namespace.Name,
					"name":      secretName,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			getResult := utils.UnmarshalCallToolResult[types.GetResourceResult](output)

			expectedGetResult := types.GetResourceResult{
				Resource: types.Resource{
					NamespacedNameResult: types.NamespacedNameResult{
						Name:      secretName,
						Namespace: fr.Namespace.Name,
					},
				},
			}

			cmpOptions := cmp.Options{
				cmpopts.IgnoreFields(types.Resource{}, "Age"),
				cmpopts.IgnoreFields(types.Resource{}, "Labels"),
				cmpopts.IgnoreFields(types.Resource{}, "Annotations"),
				cmpopts.IgnoreFields(types.Resource{}, "FormattedOutput.Data"),
			}
			Expect(cmp.Equal(getResult, expectedGetResult, cmpOptions)).To(BeTrue())
		})

		It("should get the data of a secret from a namespace using JSON output type", func() {
			By("Creating a secret with data")
			secretName := "test-secret"
			data := map[string][]byte{
				"data": []byte("test-data"),
			}
			err := kubeClient.Create(context.Background(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: fr.Namespace.Name,
				},
				Data: data,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Getting the secret")
			output, err := mcpInspector.
				MethodCall(resourceGetToolName, map[string]string{
					"group":      "",
					"version":    "v1",
					"kind":       "Secret",
					"namespace":  fr.Namespace.Name,
					"name":       secretName,
					"outputType": string(types.JSONOutputType),
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the data of the secret")
			getResult := utils.UnmarshalCallToolResult[types.GetResourceResult](output)

			fetchedJsonData := getResult.Resource.FormattedOutput.Data
			var fetchedSecret corev1.Secret
			err = json.Unmarshal([]byte(fetchedJsonData), &fetchedSecret)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedSecret.Data).To(Equal(data))
		})
	})

	Context("List Resources", func() {
		It("should list secrets from a namespace", func() {
			By("Creating 2 secrets")
			secretName1 := "test-secret-1"
			secretName2 := "test-secret-2"
			err := kubeClient.Create(context.Background(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName1,
					Namespace: fr.Namespace.Name,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			err = kubeClient.Create(context.Background(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName2,
					Namespace: fr.Namespace.Name,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			By("Listing secrets")
			output, err := mcpInspector.
				MethodCall(resourceListToolName, map[string]string{
					"group":     "",
					"version":   "v1",
					"kind":      "Secret",
					"namespace": fr.Namespace.Name,
				}).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			By("Checking the result")
			listResult := utils.UnmarshalCallToolResult[types.ListResourcesResult](output)
			Expect(listResult.Resources).To(HaveLen(2))

			expectedListResult := types.ListResourcesResult{
				Resources: []types.Resource{
					{
						NamespacedNameResult: types.NamespacedNameResult{
							Name:      secretName1,
							Namespace: fr.Namespace.Name,
						},
					},
					{
						NamespacedNameResult: types.NamespacedNameResult{
							Name:      secretName2,
							Namespace: fr.Namespace.Name,
						},
					},
				},
			}
			cmpOptions := cmp.Options{
				cmpopts.IgnoreFields(types.Resource{}, "Age"),
				cmpopts.IgnoreFields(types.Resource{}, "Labels"),
				cmpopts.IgnoreFields(types.Resource{}, "Annotations"),
				cmpopts.IgnoreFields(types.Resource{}, "FormattedOutput.Data"),
			}
			Expect(cmp.Equal(listResult, expectedListResult, cmpOptions)).To(BeTrue())
		})
	})
})
