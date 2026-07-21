# Must-gather tools

Offline tools for querying must-gather archives via [`omc`](https://github.com/gmeghnag/omc). Available with `--mode offline` or `--mode dual`.

See also: [User guide](user-guide.md) (including [which tools share common parameters](user-guide.md#common-parameters))

## Prerequisites

| Requirement | Applies to |
|-------------|------------|
| `omc` on `PATH` | All must-gather tools |
| `ovsdb-tool` on `PATH` | Database list and query tools only — those tools are **not registered** if the binary is missing |

| Tool | Description |
|------|-------------|
| [`must-gather-get-resource`](#must-gather-get-resource) | Get a specific Kubernetes resource from a must-gather archive |
| [`must-gather-list-resources`](#must-gather-list-resources) | List Kubernetes resources of a specific kind from a must-gather archive |
| [`must-gather-pod-logs`](#must-gather-pod-logs) | Get container logs from a pod in a must-gather archive |
| [`must-gather-ovnk-info`](#must-gather-ovnk-info) | Get OVN-Kubernetes networking information from a must-gather archive |
| [`must-gather-list-northbound-databases`](#must-gather-list-northbound-databases) | List OVN Northbound database files available in a must-gather archive *(needs ovsdb-tool)* |
| [`must-gather-list-southbound-databases`](#must-gather-list-southbound-databases) | List OVN Southbound database files available in a must-gather archive *(needs ovsdb-tool)* |
| [`must-gather-query-database`](#must-gather-query-database) | Query an OVN database from a must-gather archive using ovsdb-tool *(needs ovsdb-tool)* |

---

## must-gather-get-resource

Returns the resource definition in the requested format. Use this to inspect specific resource configurations, status, and metadata from the must-gather snapshot.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `must_gather_path` | string | **yes** | — | Absolute path to extracted must-gather directory |
| `kind` | string | **yes** | — | Kubernetes resource kind (e.g., Pod, Service, Node, Deployment, ConfigMap) |
| `name` | string | **yes** | — | Name of the resource to retrieve |
| `namespace` | string | no | `"default"` for namespaced resources | Namespace of the resource |
| `output_type` | string | no | table format | Output format - `"yaml"`, `"json"`, `"wide"`, or `"jsonpath"` |

### Examples

```json
{
  "must_gather_path": "/path/to/must-gather",
  "kind": "Pod",
  "namespace": "default",
  "name": "my-pod"
}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "kind": "Node",
  "name": "worker-0",
  "output_type": "yaml"
}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "kind": "ConfigMap",
  "namespace": "kube-system",
  "name": "my-config"
}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "kind": "Pod",
  "namespace": "default",
  "name": "my-pod",
  "output_type": "jsonpath='{.metadata.name}'"
}
```

---

## must-gather-list-resources

Returns a list of matching resources. Use this to discover what resources exist in the must-gather snapshot before retrieving specific ones with `must-gather-get-resource`.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `must_gather_path` | string | **yes** | — | Absolute path to extracted must-gather directory |
| `kind` | string | **yes** | — | Kubernetes resource kind (e.g., Pod, Service, Node, Deployment) |
| `namespace` | string | no | all namespaces | Filter by namespace. If omitted, lists resources across all namespaces |
| `label_selector` | string | no | — | Filter by label selector (e.g., `"app=ovnkube-node"`, `"component=network"`) |
| `output_type` | string | no | table format | Output format - `"yaml"`, `"json"`, `"wide"`, or `"jsonpath"` |

### Examples

```json
{
  "must_gather_path": "/path/to/must-gather",
  "kind": "Pod",
  "namespace": "default"
}
```

```json
{"must_gather_path": "/path/to/must-gather", "kind": "Node"}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "kind": "Pod",
  "label_selector": "app=my-app"
}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "kind": "Service",
  "namespace": "kube-system",
  "output_type": "json"
}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "kind": "Pod",
  "namespace": "default",
  "output_type": "jsonpath='{.items[*].metadata.name}'"
}
```

---

## must-gather-pod-logs

Returns log lines as an array. If neither head nor tail is specified, returns the first 100 lines by default. Use pattern matching to search for specific errors, warnings, or events in the logs.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `must_gather_path` | string | **yes** | — | Absolute path to extracted must-gather directory |
| `name` | string | **yes** | — | Name of the pod |
| `namespace` | string | no | `"default"` namespace | Namespace of the pod |
| `container` | string | no | — | Specific container name (required for multi-container pods) |
| `previous` | boolean | no | `false` | If true, get logs from previous container instance (useful for crash analysis) |
| `rotated` | boolean | no | `false` | If true, include rotated log files |

Also accepts common [`pattern`](user-guide.md#pattern-filtering) and [`head` / `tail` / `apply_tail_first`](user-guide.md#head--tail-line-limiting).

### Examples

```json
{
  "must_gather_path": "/path/to/must-gather",
  "namespace": "default",
  "name": "my-pod"
}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "namespace": "default",
  "name": "my-pod",
  "container": "my-container"
}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "namespace": "default",
  "name": "my-pod",
  "pattern": "error|Error|ERROR"
}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "namespace": "default",
  "name": "my-pod",
  "tail": 50
}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "namespace": "default",
  "name": "my-pod",
  "previous": true
}
```

---

## must-gather-ovnk-info

Use this to retrieve high-level OVN-Kubernetes networking information that helps diagnose network configuration and connectivity issues.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `must_gather_path` | string | **yes** | — | Absolute path to extracted must-gather directory |
| `info_type` | string | **yes** | — | Type of OVN-K info to retrieve. Valid values: `"extrainfo"`: Get extra OVN-Kubernetes debugging information. `"hostnetinfo"`: Get host networking configuration and status. `"subnets"`: Get OVN subnet allocations and network topology |

### Examples

```json
{"must_gather_path": "/path/to/must-gather", "info_type": "subnets"}
```

```json
{"must_gather_path": "/path/to/must-gather", "info_type": "hostnetinfo"}
```

```json
{"must_gather_path": "/path/to/must-gather", "info_type": "extrainfo"}
```

---

## must-gather-list-northbound-databases

Returns a mapping of database files to their source nodes. The Northbound database (nbdb) contains the logical network configuration: logical switches, routers, ports, ACLs, and load balancers.

Use this to discover available database files before querying with `must-gather-query-database`. Each OVN controller node may have its own database snapshot.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `must_gather_path` | string | **yes** | — | Absolute path to extracted must-gather directory |

### Examples

```json
{"must_gather_path": "/path/to/must-gather"}
```

---

## must-gather-list-southbound-databases

Returns a mapping of database files to their source nodes. The Southbound database (sbdb) contains the physical network bindings: chassis info, port bindings, MAC bindings, and datapath flows.

Use this to discover available database files before querying with `must-gather-query-database`. Each OVN controller node may have its own database snapshot.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `must_gather_path` | string | **yes** | — | Absolute path to extracted must-gather directory |

### Examples

```json
{"must_gather_path": "/path/to/must-gather"}
```

---

## must-gather-query-database

Returns query results in JSON format. Use this to inspect OVN database state for debugging network connectivity, policy enforcement, and load balancing issues.

### Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `must_gather_path` | string | **yes** | — | Absolute path to extracted must-gather directory |
| `database_name` | string | **yes** | — | Database file name from `must-gather-list-northbound-databases` or `must-gather-list-southbound-databases`. Must end with `"_nbdb"` or `"_sbdb"` |
| `table` | string | **yes** | — | OVN database table to query. Common tables: Northbound: `Logical_Switch`, `Logical_Router`, `Logical_Switch_Port`, `ACL`, `Load_Balancer`, `NAT`. Southbound: `Chassis`, `Port_Binding`, `MAC_Binding`, `Datapath_Binding`, `SB_Global` |
| `conditions` | string[] | no | returns all rows | Array of condition strings in OVSDB format: `["column","op","value"]`. Example: `["[\"hostname\",\"==\",\"worker-0\"]"]` |
| `columns` | string[] | no | returns all columns | Array of column names to return |

### Examples

```json
{
  "must_gather_path": "/path/to/must-gather",
  "database_name": "ovnkube-node-abc123_sbdb",
  "table": "Chassis"
}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "database_name": "ovnkube-node-abc123_sbdb",
  "table": "Chassis",
  "conditions": ["[\"hostname\",\"==\",\"worker-0\"]"]
}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "database_name": "ovnkube-node-abc123_nbdb",
  "table": "Logical_Switch",
  "columns": ["name", "ports"]
}
```

```json
{
  "must_gather_path": "/path/to/must-gather",
  "database_name": "ovnkube-node-abc123_sbdb",
  "table": "Port_Binding",
  "columns": ["logical_port", "chassis", "type"]
}
```
