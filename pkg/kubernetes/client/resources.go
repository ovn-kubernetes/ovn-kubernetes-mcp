package client

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// isNamespaced checks if a resource is namespaced.
func (c *OVNKMCPServerClientSet) isNamespaced(group, version, kind string) (bool, error) {
	apiResourceList, err := c.clientSet.Discovery().ServerResourcesForGroupVersion(schema.GroupVersion{Group: group, Version: version}.String())
	if err != nil {
		return false, fmt.Errorf("failed to fetch server resources for group version %s: %w", schema.GroupVersion{Group: group, Version: version}.String(), err)
	}
	for _, resource := range apiResourceList.APIResources {
		if resource.Kind == kind {
			return resource.Namespaced, nil
		}
	}
	return false, nil
}

// GetResource gets a resource by group, version, kind, name and namespace.
func (c *OVNKMCPServerClientSet) GetResource(ctx context.Context, group, version, kind, resourceName, namespace string) (*unstructured.Unstructured, error) {
	// If the namespace is not set, set it to the default namespace
	// if the resource is namespaced.
	if namespace == "" {
		isNamespaced, err := c.isNamespaced(group, version, kind)
		if err != nil {
			return nil, fmt.Errorf("failed to check if resource %s is namespaced: %w", schema.GroupVersionKind{Group: group, Version: version, Kind: kind}.String(), err)
		}
		if isNamespaced {
			namespace = metav1.NamespaceDefault
		}
	}

	gvk := schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
	restMapping, err := c.deferredDiscoveryRESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST mapping for resource %s: %w", gvk.String(), err)
	}
	return c.getResource(ctx, restMapping.Resource, resourceName, namespace)
}

// getResource gets a resource by group, version, kind, name and namespace.
func (c *OVNKMCPServerClientSet) getResource(ctx context.Context, gvr schema.GroupVersionResource, resourceName, namespace string) (*unstructured.Unstructured, error) {
	resource, err := c.dynamicClient.Resource(gvr).Namespace(namespace).
		Get(ctx, resourceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource %s with name %s and namespace %s: %w", gvr.String(), resourceName, namespace, err)
	}
	return resource, nil
}

// ListResources lists resources by group, version, kind and namespace.
func (c *OVNKMCPServerClientSet) ListResources(ctx context.Context, group, version, kind, namespace, labelSelector string) (*unstructured.UnstructuredList, error) {
	gvk := schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
	restMapping, err := c.deferredDiscoveryRESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST mapping for resource %s: %w", gvk.String(), err)
	}
	return c.listResources(ctx, restMapping.Resource, namespace, labelSelector)
}

// listResources lists resources by group, version, kind and namespace.
func (c *OVNKMCPServerClientSet) listResources(ctx context.Context, gvr schema.GroupVersionResource, namespace, labelSelector string) (*unstructured.UnstructuredList, error) {
	resources, err := c.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list resources %s with namespace %s and label selector %s: %w", gvr.String(), namespace, labelSelector, err)
	}
	return resources, nil
}
