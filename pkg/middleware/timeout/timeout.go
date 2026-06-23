package timeout

import (
	"context"
	"encoding/json"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/timeout"
)

// ToolTimeout returns an MCP receiving middleware that applies the default
// context timeout to a tools/call request only when the request does not
// carry a non-zero timeout parameter. When the tool has a non-zero timeout
// parameter, the middleware leaves the context unchanged and the tool
// handler is responsible for enforcing that per-request timeout.
func ToolTimeout(timeout time.Duration) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if method != "tools/call" {
				return next(ctx, method, req)
			}

			// Get the timeout parameters from the request.
			timeoutParam := getTimeoutParams(req)

			// If the timeout parameter is 0, apply the default timeout.
			if timeoutParam == 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}

			return next(ctx, method, req)
		}
	}
}

// getTimeoutParams extracts the timeout parameters from the request.
// If timeout parameter cannot be extracted, return 0.
func getTimeoutParams(req mcp.Request) time.Duration {
	if req == nil {
		return 0
	}
	p, ok := req.GetParams().(*mcp.CallToolParamsRaw)
	if !ok || p == nil {
		return 0
	}
	var timeoutParams timeout.TimeoutParams
	if err := json.Unmarshal(p.Arguments, &timeoutParams); err != nil {
		return 0
	}
	return timeoutParams.ToDuration()
}
