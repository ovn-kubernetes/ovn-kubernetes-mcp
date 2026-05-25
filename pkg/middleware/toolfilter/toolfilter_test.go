package toolfilter

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestToolFilter(t *testing.T) {
	t.Run("nil disabled set is a no-op for tools/list", func(t *testing.T) {
		listResult := &mcp.ListToolsResult{
			Tools: []*mcp.Tool{{Name: "tcpdump"}, {Name: "ovn-show"}},
		}
		m := ToolFilter(nil)
		h := m(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			return listResult, nil
		})

		res, err := h(context.Background(), "tools/list", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		lr := res.(*mcp.ListToolsResult)
		if len(lr.Tools) != 2 {
			t.Fatalf("expected 2 tools to pass through, got %d", len(lr.Tools))
		}
	})

	t.Run("empty disabled set is a no-op for tools/call", func(t *testing.T) {
		m := ToolFilter(map[string]bool{})
		called := false
		h := m(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			called = true
			return nil, nil
		})

		_, err := h(context.Background(), "tools/call", newCallReq("tcpdump"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("expected next handler to be called")
		}
	})

	t.Run("tools/list filters out disabled tools", func(t *testing.T) {
		listResult := &mcp.ListToolsResult{
			Tools: []*mcp.Tool{
				{Name: "tcpdump"},
				{Name: "ovn-show"},
				{Name: "pwru"},
				{Name: "ovs-list-br"},
			},
		}
		// Snapshot the original Tools slice header before invoking the
		// middleware. ToolFilter must allocate a fresh backing array for the
		// filtered output rather than reusing this one in place
		// (lr.Tools[:0] would silently mutate any caller still holding a
		// reference). Comparing this snapshot after the call catches that
		// regression.
		originalTools := listResult.Tools
		originalNames := []string{"tcpdump", "ovn-show", "pwru", "ovs-list-br"}

		m := ToolFilter(map[string]bool{"tcpdump": true, "pwru": true})
		h := m(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			return listResult, nil
		})

		res, err := h(context.Background(), "tools/list", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		lr := res.(*mcp.ListToolsResult)
		if len(lr.Tools) != 2 {
			t.Fatalf("expected 2 tools after filtering, got %d", len(lr.Tools))
		}
		for _, tool := range lr.Tools {
			if tool.Name == "tcpdump" || tool.Name == "pwru" {
				t.Fatalf("disabled tool %q was not filtered out", tool.Name)
			}
		}

		// The caller's view of the original slice must be unchanged - same
		// length, same entries in the same order. If ToolFilter ever reverts
		// to filtering in place via lr.Tools[:0], the disabled entries here
		// would be overwritten by the kept ones and this assertion fires.
		if len(originalTools) != len(originalNames) {
			t.Fatalf("caller's Tools slice length changed: got %d, want %d", len(originalTools), len(originalNames))
		}
		for i, want := range originalNames {
			if originalTools[i] == nil {
				t.Fatalf("caller's Tools[%d] was nilled out by ToolFilter", i)
			}
			if got := originalTools[i].Name; got != want {
				t.Fatalf("caller's Tools[%d].Name = %q, want %q (ToolFilter mutated the caller's backing array)", i, got, want)
			}
		}
	})

	t.Run("tools/list propagates errors from next without filtering", func(t *testing.T) {
		m := ToolFilter(map[string]bool{"tcpdump": true})
		sentinel := contextDeadline()
		h := m(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			return nil, sentinel
		})

		_, err := h(context.Background(), "tools/list", nil)
		if err != sentinel {
			t.Fatalf("expected sentinel error, got: %v", err)
		}
	})

	t.Run("tools/call rejects disabled tool with descriptive error", func(t *testing.T) {
		m := ToolFilter(map[string]bool{"tcpdump": true})
		called := false
		h := m(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			called = true
			return nil, nil
		})

		_, err := h(context.Background(), "tools/call", newCallReq("tcpdump"))
		if err == nil {
			t.Fatal("expected error rejecting disabled tool, got nil")
		}
		if called {
			t.Fatal("next handler must not be invoked for disabled tool")
		}
		if !strings.Contains(err.Error(), "tcpdump") || !strings.Contains(err.Error(), "disabled") {
			t.Fatalf("error message should mention tool name and disabled state, got: %v", err)
		}
	})

	t.Run("tools/call allows enabled tool through", func(t *testing.T) {
		m := ToolFilter(map[string]bool{"tcpdump": true})
		called := false
		h := m(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
			called = true
			return nil, nil
		})

		_, err := h(context.Background(), "tools/call", newCallReq("ovn-show"))
		if err != nil {
			t.Fatalf("unexpected error for enabled tool: %v", err)
		}
		if !called {
			t.Fatal("expected next handler to be called for enabled tool")
		}
	})

	t.Run("other methods are passed through unmodified", func(t *testing.T) {
		m := ToolFilter(map[string]bool{"tcpdump": true})
		called := false
		h := m(func(_ context.Context, method string, _ mcp.Request) (mcp.Result, error) {
			called = true
			if method != "initialize" {
				t.Fatalf("expected method 'initialize', got %q", method)
			}
			return nil, nil
		})

		if _, err := h(context.Background(), "initialize", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("expected next handler to be called for non-tool method")
		}
	})
}

// newCallReq builds a minimal CallToolRequest with the given tool name.
func newCallReq(name string) *mcp.CallToolRequest {
	return &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{Name: name},
	}
}

// contextDeadline is a sentinel error value distinct from anything the
// middleware itself produces. Used to verify error pass-through.
func contextDeadline() error { return context.DeadlineExceeded }

func TestCategories(t *testing.T) {
	cats := Categories()
	if len(cats) == 0 {
		t.Fatal("expected at least one known category")
	}
	for i := 1; i < len(cats); i++ {
		if cats[i-1] >= cats[i] {
			t.Fatalf("categories must be sorted, got %v", cats)
		}
	}
}

func TestResolveDisabled(t *testing.T) {
	t.Run("empty inputs return empty set", func(t *testing.T) {
		disabled, err := ResolveDisabled("", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(disabled) != 0 {
			t.Fatalf("expected empty set, got %v", disabled)
		}
	})

	t.Run("expands a category into its tool names", func(t *testing.T) {
		disabled, err := ResolveDisabled("network-tools", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !disabled["tcpdump"] || !disabled["pwru"] {
			t.Fatalf("expected network-tools to expand to tcpdump and pwru, got %v", disabled)
		}
	})

	t.Run("merges categories and individual tool names without duplicates", func(t *testing.T) {
		disabled, err := ResolveDisabled("network-tools", "tcpdump,ovn-show")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := map[string]bool{"tcpdump": true, "pwru": true, "ovn-show": true}
		if len(disabled) != len(want) {
			t.Fatalf("expected %d entries, got %d (%v)", len(want), len(disabled), disabled)
		}
		for name := range want {
			if !disabled[name] {
				t.Fatalf("expected %q to be disabled, got %v", name, disabled)
			}
		}
	})

	t.Run("trims whitespace and ignores empty entries", func(t *testing.T) {
		disabled, err := ResolveDisabled(" network-tools , ", " tcpdump , ,pwru ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !disabled["tcpdump"] || !disabled["pwru"] {
			t.Fatalf("expected tcpdump and pwru regardless of whitespace, got %v", disabled)
		}
	})

	t.Run("unknown category is a hard error", func(t *testing.T) {
		_, err := ResolveDisabled("does-not-exist", "")
		if err == nil {
			t.Fatal("expected error for unknown category, got nil")
		}
		if !strings.Contains(err.Error(), "does-not-exist") {
			t.Fatalf("error should mention the bad category, got: %v", err)
		}
	})

	t.Run("unknown tool name is a hard error", func(t *testing.T) {
		_, err := ResolveDisabled("", "tcpdumb")
		if err == nil {
			t.Fatal("expected error for unknown tool, got nil")
		}
		if !strings.Contains(err.Error(), "tcpdumb") {
			t.Fatalf("error should mention the bad tool name, got: %v", err)
		}
	})
}
