package client

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestDescribeResource(t *testing.T) {
	tests := []struct {
		testName       string
		objects        []runtime.Object
		gvk            schema.GroupVersionKind
		objectName     string
		namespace      string
		expectedErr    bool
		expectedEvents []string // expected event names, in the order returned
	}{
		{
			testName: "Pod resource returns only events involving it",
			objects: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx",
						Namespace: "default",
						UID:       "pod-uid",
						Labels:    map[string]string{"app": "nginx"},
					},
				},
				&corev1.Event{
					ObjectMeta: metav1.ObjectMeta{Name: "nginx.ev1", Namespace: "default"},
					InvolvedObject: corev1.ObjectReference{
						Kind:      "Pod",
						Name:      "nginx",
						Namespace: "default",
						UID:       "pod-uid",
					},
				},
				// Event for a different pod; must not be matched.
				&corev1.Event{
					ObjectMeta: metav1.ObjectMeta{Name: "other.ev1", Namespace: "default"},
					InvolvedObject: corev1.ObjectReference{
						Kind:      "Pod",
						Name:      "other-pod",
						Namespace: "default",
						UID:       "other-uid",
					},
				},
			},
			gvk:            schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
			objectName:     "nginx",
			namespace:      "default",
			expectedEvents: []string{"nginx.ev1"},
		},
		{
			testName: "Resource without matching events returns an empty event list",
			objects: []runtime.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "no-events",
						Namespace: "default",
					},
				},
			},
			gvk:            schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
			objectName:     "no-events",
			namespace:      "default",
			expectedEvents: []string{},
		},
		{
			testName: "Cluster-scoped resource (Node) matches events by UID across namespaces",
			objects: []runtime.Object{
				&corev1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "worker-0",
						UID:  "node-uid",
					},
				},
				&corev1.Event{
					ObjectMeta: metav1.ObjectMeta{Name: "worker-0.ev1", Namespace: "default"},
					InvolvedObject: corev1.ObjectReference{
						Kind: "Node",
						Name: "worker-0",
						UID:  "node-uid",
					},
				},
			},
			gvk:            schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Node"},
			objectName:     "worker-0",
			namespace:      "",
			expectedEvents: []string{"worker-0.ev1"},
		},
		{
			testName:    "Resource not found returns an error",
			objects:     []runtime.Object{},
			gvk:         schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
			objectName:  "missing",
			namespace:   "default",
			expectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			fakeclient := NewFakeClient(test.objects...)
			resource, events, err := fakeclient.DescribeResource(context.Background(), test.gvk.Group, test.gvk.Version, test.gvk.Kind, test.objectName, test.namespace)
			if (err != nil) != test.expectedErr {
				t.Fatalf("DescribeResource() error = %v, expectedErr = %v", err, test.expectedErr)
			}
			if test.expectedErr {
				return
			}

			if resource == nil {
				t.Fatalf("expected a non-nil resource")
			}
			if resource.GetName() != test.objectName {
				t.Errorf("expected resource name %q, got %q", test.objectName, resource.GetName())
			}

			if events == nil {
				t.Fatalf("expected a non-nil event list")
			}
			gotNames := make([]string, 0, len(events.Items))
			for _, e := range events.Items {
				gotNames = append(gotNames, e.Name)
			}
			if len(gotNames) != len(test.expectedEvents) {
				t.Fatalf("expected events %v, got %v", test.expectedEvents, gotNames)
			}
			for i, name := range test.expectedEvents {
				if gotNames[i] != name {
					t.Errorf("expected events %v, got %v", test.expectedEvents, gotNames)
					break
				}
			}
		})
	}
}
