package client

import (
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func convertToUnstructured(object runtime.Object, gvk schema.GroupVersionKind) (*unstructured.Unstructured, error) {
	var (
		err error
		u   unstructured.Unstructured
	)

	u.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to unstructured: %w", err)
	}

	apiv, k := gvk.ToAPIVersionAndKind()
	u.SetAPIVersion(apiv)
	u.SetKind(k)
	return &u, nil
}

func TestGetResource(t *testing.T) {
	tests := []struct {
		testName    string
		object      runtime.Object
		gvk         schema.GroupVersionKind
		objectName  string
		namespace   string
		expectedErr bool
	}{
		{
			testName: "Pod resources",
			object: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
			gvk:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
			objectName: "test",
			namespace:  "default",
		},
		{
			testName: "Node resource",
			object: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			gvk:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Node"},
			objectName: "test",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			fakeclient := NewFakeClient(test.object)
			resource, err := fakeclient.GetResource(context.Background(), test.gvk.Group, test.gvk.Version, test.gvk.Kind, test.objectName, test.namespace)
			if (err != nil) != test.expectedErr {
				t.Fatalf("Failed to get resource: %v", err)
			}
			if !test.expectedErr {
				expected, err := convertToUnstructured(test.object, test.gvk)
				if err != nil {
					t.Fatalf("Failed to convert expected object to unstructured: %v", err)
				}
				if !equality.Semantic.DeepEqual(resource, expected) {
					t.Fatalf("Resource is not equal to the expected object: \nactual: %v, \nexpected: %v", resource, expected)
				}
			}
		})
	}
}

func TestListResources(t *testing.T) {
	tests := []struct {
		testName          string
		objects           []runtime.Object
		gvk               schema.GroupVersionKind
		namespace         string
		labelSelector     string
		expectedResources []runtime.Object
		expectedErr       bool
	}{
		{
			testName: "List Pod resources in default namespace",
			objects: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "default",
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: "test-ns",
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test4",
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test5",
						Namespace: "default",
					},
				},
			},
			gvk:       schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
			namespace: "default",
			expectedResources: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "default",
					},
				},
			},
			expectedErr: false,
		},
		{
			testName: "List Pod resources in default namespace with label selector",
			objects: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
						Labels: map[string]string{
							"app": "test",
						},
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "default",
						Labels: map[string]string{
							"app": "test2",
						},
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: "test-ns",
						Labels: map[string]string{
							"app": "test3",
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test4",
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test5",
						Namespace: "default",
					},
				},
			},
			gvk:           schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
			namespace:     "default",
			labelSelector: "app=test",
			expectedResources: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
						Labels: map[string]string{
							"app": "test",
						},
					},
				},
			},
			expectedErr: false,
		},
		{
			testName: "List Pod resources across all namespaces",
			objects: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "default",
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: "test-ns",
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test4",
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test5",
						Namespace: "default",
					},
				},
			},
			namespace: "default",
			gvk:       schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
			expectedResources: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "default",
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: "test-ns",
					},
				},
			},
			expectedErr: false,
		},
		{
			testName: "List Pod resources across all namespaces with label selector",
			objects: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
						Labels: map[string]string{
							"app": "test",
						},
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "default",
						Labels: map[string]string{
							"app": "test2",
						},
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: "test-ns",
						Labels: map[string]string{
							"app": "test",
						},
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test4",
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test5",
						Namespace: "default",
					},
				},
			},
			namespace: "default",
			gvk:       schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
			expectedResources: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
						Labels: map[string]string{
							"app": "test",
						},
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: "test-ns",
						Labels: map[string]string{
							"app": "test",
						},
					},
				},
			},
			expectedErr: false,
		},
		{
			testName: "List Node resources",
			objects: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "default",
					},
				},
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test3",
					},
				},
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test4",
						Namespace: "default",
					},
				},
			},
			gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Node"},
			expectedResources: []runtime.Object{
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test3",
					},
				},
			},
			expectedErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			fakeclient := NewFakeClient(test.objects...)
			resources, err := fakeclient.ListResources(context.Background(), test.gvk.Group, test.gvk.Version, test.gvk.Kind, test.namespace, test.labelSelector)
			if (err != nil) != test.expectedErr {
				t.Fatalf("Failed to list resources: %v", err)
			}
			if !test.expectedErr {
				expectedResources := make([]unstructured.Unstructured, 0)
				for _, resource := range resources.Items {
					expected, err := convertToUnstructured(&resource, test.gvk)
					if err != nil {
						t.Fatalf("Failed to convert expected object to unstructured: %v", err)
					}
					expectedResources = append(expectedResources, *expected)
				}
				if !equality.Semantic.DeepEqual(resources.Items, expectedResources) {
					t.Fatalf("Resources are not equal to the expected objects: \nactual: %v, \nexpected: %v", resources.Items, expectedResources)
				}
			}
		})
	}
}
