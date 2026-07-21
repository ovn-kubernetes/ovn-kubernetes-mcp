# Kubernetes tools

Live-cluster tools for inspecting pods and Kubernetes API resources. Available with `--mode live-cluster` or `--mode dual`.

See also: [User guide](user-guide.md) (including [common parameters](user-guide.md#common-parameters))

| Tool | Description |
|------|-------------|
| [`pod-logs`](#pod-logs) | Get container logs from a pod in the Kubernetes cluster |
| [`resource-get`](#resource-get) | Get a specific Kubernetes resource by name |
| [`resource-list`](#resource-list) | List Kubernetes resources of a specific kind |

---

## pod-logs

Retrieves logs from a running or terminated pod. Supports filtering by pattern, limiting output with head/tail, and retrieving logs from previous container instances.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `name` | string | **yes** | — | Name of the pod |
| `namespace` | string | no | `"default"` | Namespace of the pod |
| `container` | string | no | — | Specific container name (required for multi-container pods) |
| `previous` | boolean | no | `false` | If true, get logs from previous container instance (useful for crash analysis) |

Also accepts common [`pattern`](user-guide.md#pattern-filtering) and [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting).

### Examples

```json
{"name": "my-pod", "namespace": "default"}
```

```json
{"name": "my-pod", "namespace": "default", "container": "my-container"}
```

```json
{"name": "my-pod", "namespace": "default", "previous": true}
```

```json
{"name": "my-pod", "namespace": "default", "pattern": "error|Error|ERROR"}
```

```json
{"name": "my-pod", "namespace": "default", "tail": 50}
```

```json
{
  "name": "my-pod",
  "namespace": "default",
  "head": 50,
  "tail": 100,
  "apply_tail_first": true
}
```

---

## resource-get

Retrieves a single resource from the cluster using its group, version, kind, and name. Supports different output formats for viewing the resource data.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `group` | string | no | empty (core resources) | API group of the resource (e.g., `"apps"`, `"networking.k8s.io"`). Empty for core resources |
| `version` | string | **yes** | — | API version of the resource (e.g., `"v1"`, `"v1beta1"`) |
| `kind` | string | **yes** | — | Kind of the resource (e.g., `"Pod"`, `"Service"`, `"Deployment"`, `"ConfigMap"`) |
| `name` | string | **yes** | — | Name of the resource to retrieve |
| `namespace` | string | no | `"default"` for namespaced resources; empty for cluster-scoped | Namespace of the resource. If omitted, defaults to `"default"` for namespaced resources and empty for cluster-scoped resources |
| `output_type` | string | no | table format with name, namespace, age | Output format - `"yaml"`, `"json"`, `"jsonpath"`, or `"wide"` (wide will include labels and annotations) |

### Examples

```json
{"version": "v1", "kind": "Pod", "name": "my-pod"}
```

```json
{
  "group": "apps",
  "version": "v1",
  "kind": "Deployment",
  "name": "my-deployment",
  "namespace": "default",
  "output_type": "yaml"
}
```

```json
{"version": "v1", "kind": "Node", "name": "worker-0"}
```

```json
{
  "version": "v1",
  "kind": "ConfigMap",
  "name": "my-config",
  "namespace": "kube-system",
  "output_type": "json"
}
```

```json
{
  "version": "v1",
  "kind": "Pod",
  "name": "my-pod",
  "namespace": "default",
  "output_type": "jsonpath='{.metadata.name}'"
}
```

```json
{
  "version": "v1",
  "kind": "Pod",
  "name": "my-pod",
  "namespace": "default",
  "output_type": "wide"
}
```

---

## resource-list

Lists resources in a namespace or across all namespaces. Supports filtering by label selector and different output formats.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `group` | string | no | empty (core resources) | API group of the resource (e.g., `"apps"`, `"networking.k8s.io"`). Empty for core resources |
| `version` | string | **yes** | — | API version of the resource (e.g., `"v1"`, `"v1beta1"`) |
| `kind` | string | **yes** | — | Kind of the resource (e.g., `"Pod"`, `"Service"`, `"Deployment"`) |
| `namespace` | string | no | all namespaces | Filter by namespace. If omitted, lists resources across all namespaces |
| `label_selector` | string | no | — | Filter by label selector (e.g., `"app=my-app"`, `"component=network"`) |
| `output_type` | string | no | table format with name, namespace, age | Output format - `"yaml"`, `"json"`, `"jsonpath"`, or `"wide"` (wide will include labels and annotations) |

### Examples

```json
{"version": "v1", "kind": "Pod", "namespace": "default"}
```

```json
{"version": "v1", "kind": "Service"}
```

```json
{
  "version": "v1",
  "kind": "Pod",
  "namespace": "default",
  "label_selector": "app=my-app"
}
```

```json
{
  "group": "apps",
  "version": "v1",
  "kind": "Deployment",
  "namespace": "default",
  "output_type": "yaml"
}
```

```json
{"version": "v1", "kind": "Node"}
```

```json
{
  "version": "v1",
  "kind": "Pod",
  "namespace": "kube-system",
  "output_type": "wide"
}
```

```json
{
  "version": "v1",
  "kind": "Pod",
  "namespace": "default",
  "output_type": "jsonpath='{.items[*].metadata.name}'"
}
```
