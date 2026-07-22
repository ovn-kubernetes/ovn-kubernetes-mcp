package types

import (
	"bytes"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kubectldescribe "k8s.io/kubectl/pkg/describe"
	yaml "sigs.k8s.io/yaml"
)

// GetResourceParams is a type that contains the group, version, kind, name and namespace of a resource.
type GetResourceParams struct {
	GroupVersionKind
	GetParams
}

// GetResourceResult is a type that contains the resource data.
type GetResourceResult struct {
	Resource Resource `json:"resource"`
}

// ListParams is a type that contains the namespace, label selector and output type of a resource.
type ListParams struct {
	Namespace     string `json:"namespace,omitempty"`
	LabelSelector string `json:"label_selector,omitempty"`
	OutputParams
}

// ListResourcesParams is a type that contains the group, version, kind, namespace and output type of a resource.
type ListResourcesParams struct {
	GroupVersionKind
	ListParams
}

// ListResourcesResult is a type that contains the resource data.
type ListResourcesResult struct {
	Resources []Resource `json:"resources"`
}

// DescribeResourceParams is a type that contains the group, version, kind, name and
// namespace of a resource to describe.
type DescribeResourceParams struct {
	GroupVersionKind
	NamespacedNameParams
}

// DescribeResourceResult is a type that contains the human-readable description of a
// resource, including its spec, status and related events, similar to `kubectl describe`.
type DescribeResourceResult struct {
	Description string `json:"description"`
}

// Describe renders the resource's identifying metadata, labels, annotations, spec,
// status and events into r.Description as a human-readable string, similar to
// `kubectl describe` output.
func (r *DescribeResourceResult) Describe(resource *unstructured.Unstructured, events *corev1.EventList) {
	buf := &bytes.Buffer{}
	w := kubectldescribe.NewPrefixWriter(buf)

	// writeMap writes a "<title>:" line with the first key=value pair on the
	// header line and the rest indented below it, matching `kubectl describe`.
	writeMap := func(title string, m map[string]string) {
		if len(m) == 0 {
			w.Write(kubectldescribe.LEVEL_0, "%s:\t<none>\n", title)
			return
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		w.Write(kubectldescribe.LEVEL_0, "%s:\t%s=%s\n", title, keys[0], m[keys[0]])
		for _, k := range keys[1:] {
			w.Write(kubectldescribe.LEVEL_1, "%s=%s\n", k, m[k])
		}
	}

	// writeYAML writes a "<title>:" section with data rendered as indented YAML.
	writeYAML := func(title string, data any) {
		yamlBytes, err := yaml.Marshal(data)
		if err != nil {
			w.Write(kubectldescribe.LEVEL_0, "%s:\t<error rendering: %v>\n", title, err)
			return
		}
		trimmed := strings.TrimRight(string(yamlBytes), "\n")
		if trimmed == "" || trimmed == "{}" {
			w.Write(kubectldescribe.LEVEL_0, "%s:\t<none>\n", title)
			return
		}
		w.Write(kubectldescribe.LEVEL_0, "%s:\n", title)
		for _, line := range strings.Split(trimmed, "\n") {
			w.Write(kubectldescribe.LEVEL_1, "%s\n", line)
		}
	}

	w.Write(kubectldescribe.LEVEL_0, "Name:\t%s\n", resource.GetName())
	if ns := resource.GetNamespace(); ns != "" {
		w.Write(kubectldescribe.LEVEL_0, "Namespace:\t%s\n", ns)
	}
	w.Write(kubectldescribe.LEVEL_0, "Kind:\t%s\n", resource.GetKind())
	if apiVersion := resource.GetAPIVersion(); apiVersion != "" {
		w.Write(kubectldescribe.LEVEL_0, "API Version:\t%s\n", apiVersion)
	}
	writeMap("Labels", resource.GetLabels())
	writeMap("Annotations", resource.GetAnnotations())

	if spec, found, _ := unstructured.NestedFieldNoCopy(resource.Object, "spec"); found {
		writeYAML("Spec", spec)
	}
	if status, found, _ := unstructured.NestedFieldNoCopy(resource.Object, "status"); found {
		writeYAML("Status", status)
	}

	w.Write(kubectldescribe.LEVEL_0, "\n")
	kubectldescribe.DescribeEvents(events, w)
	w.Flush()

	r.Description = buf.String()
}
