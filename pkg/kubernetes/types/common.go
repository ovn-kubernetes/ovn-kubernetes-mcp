package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/jsonpath"
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

// ToJSONPath gets the JSONPath data from a resource.
func (j *FormattedOutput) ToJSONPath(template string, data map[string]any) error {
	jp := jsonpath.New("jsonpath")
	if err := jp.Parse(template); err != nil {
		return err
	}
	dataBuffer := bytes.NewBuffer(nil)
	err := jp.Execute(dataBuffer, data)
	if err != nil {
		return fmt.Errorf("failed to execute jsonpath template %s, error: %w", template, err)
	}
	j.Data = dataBuffer.String()
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
	// JSONPathOutputType is the output type for JSONPath data.
	JSONPathOutputType OutputType = "jsonpath"
	// WideOutputType is the output type for detailed data.
	WideOutputType OutputType = "wide"
)

// OutputParams is a type that contains the output type and JSONPathTemplate of a resource.
type OutputParams struct {
	// OutputType is the output type of the resource. If set, it can be YAML, JSON, JSONPath or wide.
	OutputType OutputType `json:"output_type,omitempty"`
	// JSONPathTemplate is the JSONPathTemplate template to use for the output.
	JSONPathTemplate string `json:"json_path_template,omitempty"`
}

// ValidateOutputParams validates the output parameters.
func (o *OutputParams) ValidateOutputParams() error {
	if o.OutputType == JSONPathOutputType && o.JSONPathTemplate == "" {
		return fmt.Errorf("json_path_template is required when output_type is jsonpath")
	}
	if o.OutputType == JSONPathOutputType {
		err := jsonpath.NewParser("validate").Parse(o.JSONPathTemplate)
		if err != nil {
			return fmt.Errorf("invalid json_path_template: %s, error: %w", o.JSONPathTemplate, err)
		}
	}
	return nil
}

// GetParams is a type that contains the name, namespace and output type of a resource.
type GetParams struct {
	NamespacedNameParams
	OutputParams
}
