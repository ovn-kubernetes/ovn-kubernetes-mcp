package toolfilter

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// toolsByCategory enumerates the tools registered by each pkg/<category>/mcp
// package. It is the source of truth for resolving disabled categories into
// individual tool names, and for typo-detecting tool names supplied by the
// operator. Keep this in sync with the AddTools implementations under
// pkg/*/mcp/mcp.go; the README "Tools available in MCP Server" table is the
// canonical reference.
var toolsByCategory = map[string][]string{
	"kubernetes":    {"pod-logs", "resource-get", "resource-list"},
	"ovn":           {"ovn-show", "ovn-get", "ovn-lflow-list", "ovn-trace"},
	"ovs":           {"ovs-list-br", "ovs-list-ports", "ovs-list-ifaces", "ovs-vsctl-show", "ovs-ofctl-dump-flows", "ovs-appctl-dump-conntrack", "ovs-appctl-ofproto-trace"},
	"kernel":        {"get-conntrack", "get-iptables", "get-nft", "get-ip"},
	"network-tools": {"tcpdump", "pwru"},
	"sosreport":     {"sos-list-plugins", "sos-list-commands", "sos-search-commands", "sos-get-command", "sos-search-pod-logs"},
	"must-gather":   {"must-gather-get-resource", "must-gather-list-resources", "must-gather-pod-logs", "must-gather-ovnk-info", "must-gather-list-northbound-databases", "must-gather-list-southbound-databases", "must-gather-query-database"},
}

// Categories returns a sorted list of known tool category names. Intended for
// building CLI help strings and error messages.
func Categories() []string {
	return slices.Sorted(maps.Keys(toolsByCategory))
}

// ResolveDisabled expands comma-separated category and tool name lists into a
// single set of tool names to hide. Whitespace in either list is trimmed and
// empty entries are ignored. Unknown category or tool names return an error so
// configuration bugs (typos, stale configs) fail fast at startup rather than
// silently disabling nothing.
func ResolveDisabled(disabledCategories, disabledTools string) (map[string]bool, error) {
	disabled := map[string]bool{}

	for _, cat := range splitCSV(disabledCategories) {
		tools, ok := toolsByCategory[cat]
		if !ok {
			return nil, fmt.Errorf("unknown category %q (valid: %s)", cat, strings.Join(Categories(), ", "))
		}
		for _, t := range tools {
			disabled[t] = true
		}
	}

	known := allKnownTools()
	for _, t := range splitCSV(disabledTools) {
		if !known[t] {
			return nil, fmt.Errorf("unknown tool %q", t)
		}
		disabled[t] = true
	}

	return disabled, nil
}

// ToolFilter returns an MCP receiving middleware that hides tools whose names
// appear in disabledTools from tools/list responses and rejects tools/call
// requests targeting those names with a "disabled by server configuration"
// error. A nil or empty disabledTools map makes the middleware a no-op, so it
// is safe to install unconditionally.
//
// Filtering at the middleware layer keeps every pkg/<category>/mcp registration
// site untouched: tools are still registered with the SDK, the server's
// listTools handler still returns them, and this wrapper strips them out on the
// way back to the client.
func ToolFilter(disabledTools map[string]bool) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if len(disabledTools) == 0 {
				return next(ctx, method, req)
			}

			switch method {
			case "tools/call":
				if name := callToolName(req); name != "" && disabledTools[name] {
					return nil, fmt.Errorf("tool %q is disabled by server configuration", name)
				}
				return next(ctx, method, req)

			case "tools/list":
				res, err := next(ctx, method, req)
				if err != nil {
					return res, err
				}
				if lr, ok := res.(*mcp.ListToolsResult); ok && len(lr.Tools) > 0 {
					filtered := make([]*mcp.Tool, 0, len(lr.Tools))
					for _, t := range lr.Tools {
						if t == nil || disabledTools[t.Name] {
							continue
						}
						filtered = append(filtered, t)
					}
					lr.Tools = filtered
				}
				return res, err

			default:
				return next(ctx, method, req)
			}
		}
	}
}

// splitCSV splits a comma-separated string, trims whitespace from each entry,
// and drops empties so " a, ,b " becomes []string{"a","b"}.
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := parts[:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// allKnownTools returns the union of every tool name in toolsByCategory as a
// set, used for typo detection on disabled tool names.
func allKnownTools() map[string]bool {
	out := map[string]bool{}
	for _, tools := range toolsByCategory {
		for _, t := range tools {
			out[t] = true
		}
	}
	return out
}

// callToolName extracts the tool name from a tools/call request. Returns "" if
// the request shape is unexpected.
func callToolName(req mcp.Request) string {
	if req == nil {
		return ""
	}
	p, ok := req.GetParams().(*mcp.CallToolParamsRaw)
	if !ok || p == nil {
		return ""
	}
	return p.Name
}
