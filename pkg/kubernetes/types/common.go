package types

import (
	"encoding/json"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yaml "sigs.k8s.io/yaml"
)

// FormattedOutput is a type that contains the formatted data of a resource.
type FormattedOutput struct {
	Data string `json:"data"`
}

// ToJSON gets the JSON data from a resource.
func (j *FormattedOutput) ToJSON(data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	j.Data = string(jsonData)
	return nil
}

// ToYAML gets the YAML data from a resource.
func (j *FormattedOutput) ToYAML(data any) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	j.Data = string(yamlData)
	return nil
}

// NamespacedNameParams is a type that contains the name and namespace of a resource.
type NamespacedNameParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// NamespacedNameResult is a type that contains the name and namespace of a resource.
// The fields are optional.
type NamespacedNameResult struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// Resource is a type that contains the name, namespace, age, labels and annotations of a resource.
type Resource struct {
	NamespacedNameResult
	Age         string            `json:"age,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	FormattedOutput
}

// GetResourceData gets the data of a resource. If isDetailed is true, the labels and annotations are also included.
func (r *Resource) GetResourceData(resource *unstructured.Unstructured, isDetailed bool) {
	r.Name = resource.GetName()
	r.Namespace = resource.GetNamespace()
	r.Age = FormatAge(time.Since(resource.GetCreationTimestamp().Time))

	if isDetailed {
		r.Labels = resource.GetLabels()
		r.Annotations = resource.GetAnnotations()
	}
}

// GroupVersionKind is a type that contains the group, version and kind of a resource.
type GroupVersionKind struct {
	Group   string `json:"group,omitempty"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

// OutputType is a type that contains the output type of a resource.
type OutputType string

const (
	// YAMLOutputType is the output type for YAML data.
	YAMLOutputType OutputType = "yaml"
	// JSONOutputType is the output type for JSON data.
	JSONOutputType OutputType = "json"
	// WideOutputType is the output type for detailed data.
	WideOutputType OutputType = "wide"
)

// GetParams is a type that contains the name, namespace and output type of a resource.
type GetParams struct {
	NamespacedNameParams
	// OutputType is the output type of the resource. If set, it can be YAML, JSON or wide.
	OutputType OutputType `json:"output_type,omitempty"`
}
