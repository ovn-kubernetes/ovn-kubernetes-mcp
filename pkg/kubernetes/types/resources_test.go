package types

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestDescribeResourceResultDescribe(t *testing.T) {
	tests := []struct {
		testName        string
		resource        *unstructured.Unstructured
		events          *corev1.EventList
		expectedContain []string
		expectedOmit    []string
	}{
		{
			testName: "namespaced resource with labels, spec, status and events",
			resource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]any{
						"name":      "nginx",
						"namespace": "default",
						"labels":    map[string]any{"app": "nginx"},
					},
					"spec": map[string]any{
						"nodeName": "worker-1",
					},
					"status": map[string]any{
						"phase": "Running",
						"podIP": "10.244.1.5",
					},
				},
			},
			events: &corev1.EventList{
				Items: []corev1.Event{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "ev1"},
						Type:       "Normal",
						Reason:     "Scheduled",
						Message:    "Successfully assigned default/nginx to worker-1",
						Source:     corev1.EventSource{Component: "default-scheduler"},
					},
				},
			},
			expectedContain: []string{
				"Name:\tnginx",
				"Namespace:\tdefault",
				"Kind:\tPod",
				"API Version:\tv1",
				"Labels:\tapp=nginx",
				"Annotations:\t<none>",
				"Spec:",
				"nodeName: worker-1",
				"Status:",
				"phase: Running",
				"podIP: 10.244.1.5",
				"Events:",
				"Scheduled",
				"Successfully assigned default/nginx to worker-1",
			},
		},
		{
			testName: "resource without labels, spec, status or events shows <none>",
			resource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]any{
						"name":      "empty",
						"namespace": "default",
					},
				},
			},
			events: &corev1.EventList{},
			expectedContain: []string{
				"Name:\tempty",
				"Labels:\t<none>",
				"Annotations:\t<none>",
				"Events:\t<none>",
			},
			expectedOmit: []string{"Spec:", "Status:"},
		},
		{
			testName: "cluster-scoped resource omits Namespace line",
			resource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "v1",
					"kind":       "Node",
					"metadata": map[string]any{
						"name": "worker-0",
					},
				},
			},
			events: &corev1.EventList{},
			expectedContain: []string{
				"Name:\tworker-0",
				"Kind:\tNode",
			},
			expectedOmit: []string{"Namespace:\t"},
		},
		{
			testName: "multiple labels are sorted, first on the header line",
			resource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]any{
						"name": "multi-label",
						"labels": map[string]any{
							"zeta":  "z",
							"alpha": "a",
						},
					},
				},
			},
			events: &corev1.EventList{},
			expectedContain: []string{
				"Labels:\talpha=a",
				"zeta=z",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			result := DescribeResourceResult{}
			result.Describe(test.resource, test.events)
			for _, want := range test.expectedContain {
				if !strings.Contains(result.Description, want) {
					t.Errorf("expected description to contain %q, got:\n%s", want, result.Description)
				}
			}
			for _, notWant := range test.expectedOmit {
				if strings.Contains(result.Description, notWant) {
					t.Errorf("expected description to not contain %q, got:\n%s", notWant, result.Description)
				}
			}
		})
	}
}
