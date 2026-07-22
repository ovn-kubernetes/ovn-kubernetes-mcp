package client

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// DescribeResource fetches a resource by group, version, kind, name and namespace
// together with the Kubernetes Events associated with it. Unlike GetResource, it
// also returns the resource's Events so that callers can render a description
// similar to `kubectl describe`. It works generically for any resource kind
// (built-in or CRD) since it relies only on the dynamic client and the resource's
// metadata to correlate events.
func (c *OVNKMCPServerClientSet) DescribeResource(ctx context.Context, group, version, kind, name, namespace string) (*unstructured.Unstructured, *corev1.EventList, error) {
	resource, err := c.GetResource(ctx, group, version, kind, name, namespace)
	if err != nil {
		return nil, nil, err
	}

	events, err := c.getEventsForObject(ctx, resource)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get events for resource %s with name %s: %w", kind, name, err)
	}

	return resource, events, nil
}

// getEventsForObject returns the Events associated with the given object. Objects
// are matched by UID when available, falling back to namespace/name/kind so that
// events are still surfaced for objects whose UID was stripped (e.g. by a cache).
func (c *OVNKMCPServerClientSet) getEventsForObject(ctx context.Context, object *unstructured.Unstructured) (*corev1.EventList, error) {
	allEvents, err := c.clientSet.CoreV1().Events(object.GetNamespace()).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	matched := &corev1.EventList{}
	for _, event := range allEvents.Items {
		if eventMatchesObject(event, object) {
			matched.Items = append(matched.Items, event)
		}
	}
	return matched, nil
}

// eventMatchesObject reports whether the event's InvolvedObject reference points
// at the given object.
func eventMatchesObject(event corev1.Event, object *unstructured.Unstructured) bool {
	if object.GetUID() != "" && event.InvolvedObject.UID == object.GetUID() {
		return true
	}
	return event.InvolvedObject.Name == object.GetName() &&
		event.InvolvedObject.Namespace == object.GetNamespace() &&
		event.InvolvedObject.Kind == object.GetKind()
}
