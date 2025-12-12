# ovn-kubernetes-mcp
Repo hosting the Model Context Protocol Server for troubleshooting OVN-Kubernetes

## How to connect to the MCP Server

For connecting to the MCP server, the following steps are required:

```shell
make build
```

The server supports 3 operating modes:
- `live-cluster` (default): Connect to a live Kubernetes cluster for real-time debugging
- `offline`: Offline troubleshooting without requiring cluster access
- `both`: In this mode, tools from both `live-cluster` and `offline` modes will be available.

The server currently supports 2 transport modes: `stdio` and `http`.

### Live Cluster Mode

For `stdio` mode, the server can be run and connected to by using the following configuration in an MCP host (Cursor, Claude, etc.):

```json
{
  "mcpServers": {
    "ovn-kubernetes": {
      "command": "/PATH-TO-THE-LOCAL-GIT-REPO/_output/ovnk-mcp-server",
      "args": [
        "--kubeconfig",
        "/PATH-TO-THE-KUBECONFIG-FILE"
      ]
    }
  }
}
```

For `http` mode, the server should be started separately:

```shell
./PATH-TO-THE-LOCAL-GIT-REPO/_output/ovnk-mcp-server --transport http --kubeconfig /PATH-TO-THE-KUBECONFIG-FILE
```

The following configuration should be used in an MCP host (Cursor, Claude, etc.) to connect to the server:

```json
{
  "mcpServers": {
    "ovn-kubernetes": {
      "url": "http://localhost:8080"
    }
  }
}
```

### Offline Mode

For offline troubleshooting (e.g., analyzing sosreports, must-gathers), use `--mode offline`:

For `stdio` mode:

```json
{
  "mcpServers": {
    "ovn-kubernetes": {
      "command": "/PATH-TO-THE-LOCAL-GIT-REPO/_output/ovnk-mcp-server",
      "args": ["--mode", "offline"]
    }
  }
}
```

For `http` mode:

```shell
./PATH-TO-THE-LOCAL-GIT-REPO/_output/ovnk-mcp-server --transport http --mode offline
```

> NOTE: For must-gather tools to be added, availability of [`omc`](https://github.com/gmeghnag/omc) binary is mandatory. Otherwise, the must-gather tools will not be added to the MCP server. Additionally, if [`ovsdb-tool`](https://www.openvswitch.org/support/dist-docs/ovsdb-tool.1.txt) binary is not available, then some of the must-gather tools, which use `ovsdb-tool` binary, will not be added to the MCP server.

### Both Mode

For using both [live-cluster](#live-cluster-mode) (needs kubeconfig) and [offline](#offline-mode) tools, use `--mode both`:

For `stdio` mode:

```json
{
  "mcpServers": {
    "ovn-kubernetes": {
      "command": "/PATH-TO-THE-LOCAL-GIT-REPO/_output/ovnk-mcp-server",
      "args": [
        "--mode",
        "both",
        "--kubeconfig",
        "/PATH-TO-THE-KUBECONFIG-FILE"
      ]
    }
  }
}
```

For `http` mode:

```shell
./PATH-TO-THE-LOCAL-GIT-REPO/_output/ovnk-mcp-server --transport http --mode both --kubeconfig /PATH-TO-THE-KUBECONFIG-FILE
```
