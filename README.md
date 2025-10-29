# ovn-kubernetes-mcp
Repo hosting the Model Context Protocol Server for troubleshooting OVN-Kubernetes

## How to connect to the MCP Server

For connecting to the MCP server, the following steps are required:

```shell
make build
```

The server supports 2 operating modes:
- `live-cluster` (default): Connect to a live Kubernetes cluster for real-time debugging
- `offline`: Offline troubleshooting without requiring cluster access

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

For offline troubleshooting (e.g., analyzing sosreports), use `--mode offline`:

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
