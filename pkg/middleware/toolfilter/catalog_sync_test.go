package toolfilter

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/mcptools"
)

// TestToolsByCategorySync verifies that the toolsByCategory map matches the
// tools actually registered via mcp.AddTool calls under pkg/*/mcp/mcp.go.
// If a contributor adds, removes, or renames a tool without updating the
// map, this test fails with a diff naming the exact category and tool change.
func TestToolsByCategorySync(t *testing.T) {
	repoRoot, err := mcptools.FindRepoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	catalog, err := mcptools.Catalog(repoRoot)
	if err != nil {
		t.Fatalf("scan pkg/*/mcp/mcp.go: %v", err)
	}

	fromCode := make(map[string][]string, len(catalog))
	for cat, tools := range catalog {
		names := make([]string, 0, len(tools))
		for _, tool := range tools {
			names = append(names, tool.Name)
		}
		fromCode[cat] = names
	}

	if diff := cmp.Diff(toolsByCategory, fromCode); diff != "" {
		t.Errorf("toolsByCategory must match pkg/*/mcp/mcp.go (-map, +code):\n%s", diff)
	}
}
