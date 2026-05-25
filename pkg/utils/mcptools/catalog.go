// Package mcptools holds helpers that parse the MCP tool registrations under
// pkg/*/mcp/mcp.go. The same primitives are consumed by two places that must
// agree about which tools exist:
//
//   - hack/gen-readme-tools.go renders the "Tools available in MCP Server"
//     table in README.md from these registrations.
//   - pkg/middleware/toolfilter/catalog_sync_test.go asserts the
//     toolsByCategory map in pkg/middleware/toolfilter/toolfilter.go matches
//     the live registrations, so the --disable-categories / --disable-tools
//     flags cannot silently drift away from what the server actually exposes.
//
// Keeping the AST scanner in one place eliminates the risk of those two
// callers disagreeing about how a tool declaration is recognised.
package mcptools

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	// formatVerbSentinel is used while transforming format strings so a literal
	// "%%" survives a pass through format-verb stripping. Multi-rune so it cannot
	// collide with anything that occurs in real descriptions.
	formatVerbSentinel = "\uE000\uE001"

	// sdkImportPath is the canonical Go import path of the MCP SDK. Files that
	// do not import this path are assumed not to register any MCP tools.
	sdkImportPath = "github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	reFirstSentence = regexp.MustCompile(`^(.*?\.\s)`)
	rePeriodAtEnd   = regexp.MustCompile(`^(.*\.)$`)
	reFormatVerb    = regexp.MustCompile(`%[0-9.*+# \-]*[a-zA-Z]`)
)

// Tool represents a single MCP tool extracted from an mcp.AddTool registration
// in a pkg/<category>/mcp/mcp.go file. Description is shortened to the first
// line / first sentence so it renders inline in a Markdown table.
type Tool struct {
	Name        string
	Description string
}

// FindRepoRoot returns the directory that contains go.mod, walking up from the
// current working directory. Both the README generator and the sync test rely
// on this so they keep working regardless of which subdirectory they were
// invoked from.
func FindRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// ExtractTools parses the given mcp.go file and returns every MCP tool
// registered via <alias>.AddTool(server, &<alias>.Tool{...}, ...). The local
// alias of the MCP SDK is detected from the file's imports, so a file that
// uses, say, `import sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"` is
// still scanned correctly. Tools are returned in source order so callers that
// care about registration order (the README generator, the sync test) get a
// stable, meaningful sequence.
func ExtractTools(mcpPath string) ([]Tool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, mcpPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	mcpAlias, dotImported, ok := findMCPImportAlias(node)
	if !ok {
		// File doesn't import the MCP SDK (or imports it as `_`). No tool
		// registrations are possible.
		return nil, nil
	}
	var tools []Tool
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}
		if !isMCPAddToolCall(call.Fun, mcpAlias, dotImported) {
			return true
		}
		unary, ok := call.Args[1].(*ast.UnaryExpr)
		if !ok || unary.Op != token.AND {
			return true
		}
		lit, ok := unary.X.(*ast.CompositeLit)
		if !ok {
			return true
		}
		var name, desc string
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			keyIdent, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}
			switch keyIdent.Name {
			case "Name":
				if bl, ok := kv.Value.(*ast.BasicLit); ok && bl.Kind == token.STRING {
					name = UnquoteString(bl.Value)
				}
			case "Description":
				desc = extractDescription(kv.Value)
			}
		}
		if name != "" {
			tools = append(tools, Tool{Name: name, Description: firstLine(desc)})
		}
		return true
	})
	return tools, nil
}

// findMCPImportAlias scans file.Imports for the MCP SDK and returns the local
// alias used to qualify references to it. The bool result is false when the
// file doesn't import the SDK at all (or imports it blank, with "_"), in which
// case the caller should skip the file. When the SDK is dot-imported, alias is
// empty and dotImported is true so the caller can recognise bare identifiers.
func findMCPImportAlias(file *ast.File) (alias string, dotImported, ok bool) {
	for _, imp := range file.Imports {
		if imp.Path == nil || UnquoteString(imp.Path.Value) != sdkImportPath {
			continue
		}
		if imp.Name == nil {
			// Unaliased: local name is the import path's last segment, which
			// is "mcp" for this SDK.
			return "mcp", false, true
		}
		switch imp.Name.Name {
		case ".":
			return "", true, true
		case "_":
			return "", false, false
		default:
			return imp.Name.Name, false, true
		}
	}
	return "", false, false
}

// isMCPAddToolCall reports whether fun refers to the SDK's AddTool function,
// taking into account the local alias the file uses (or dot-imports).
func isMCPAddToolCall(fun ast.Expr, mcpAlias string, dotImported bool) bool {
	if dotImported {
		ident, ok := fun.(*ast.Ident)
		return ok && ident.Name == "AddTool"
	}
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "AddTool" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == mcpAlias
}

// Catalog walks every pkg/*/mcp/mcp.go file under repoRoot and returns a map
// of category name (the pkg/<name> directory) to the tools registered there
// in source order. Subdirectories that contain no mcp/mcp.go file, and files
// that register zero tools, are silently skipped.
func Catalog(repoRoot string) (map[string][]Tool, error) {
	pkgDir := filepath.Join(repoRoot, "pkg")
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", pkgDir, err)
	}
	out := make(map[string][]Tool)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mcpPath := filepath.Join(pkgDir, e.Name(), "mcp", "mcp.go")
		if _, err := os.Stat(mcpPath); os.IsNotExist(err) {
			continue
		}
		tools, err := ExtractTools(mcpPath)
		if err != nil {
			return nil, fmt.Errorf("extract %s: %w", mcpPath, err)
		}
		if len(tools) == 0 {
			continue
		}
		out[e.Name()] = tools
	}
	return out, nil
}

// UnquoteString removes surrounding backticks or double quotes from a Go
// string literal. It uses strconv.Unquote to handle escape sequences inside
// double-quoted strings and falls back to a literal strip if that fails.
func UnquoteString(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '`' && s[len(s)-1] == '`' {
		return s[1 : len(s)-1]
	}
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"') {
		if u, err := strconv.Unquote(s); err == nil {
			return u
		}
		return s[1 : len(s)-1]
	}
	return s
}

// extractDescription returns the string value of a Tool's Description field
// from the AST. It handles both plain string literals and
// fmt.Sprintf(format, ...) calls; format verbs in the latter are stripped so
// they don't appear literally in the README.
func extractDescription(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.BasicLit:
		if v.Kind == token.STRING {
			return UnquoteString(v.Value)
		}
	case *ast.CallExpr:
		if len(v.Args) >= 1 {
			if bl, ok := v.Args[0].(*ast.BasicLit); ok && bl.Kind == token.STRING {
				return stripFormatVerbs(UnquoteString(bl.Value))
			}
		}
	}
	return ""
}

// stripFormatVerbs replaces Go format verbs (e.g. %d, %s) with placeholders so
// format strings extracted from fmt.Sprintf read naturally in the README.
func stripFormatVerbs(s string) string {
	s = strings.ReplaceAll(s, "%%", formatVerbSentinel)
	s = reFormatVerb.ReplaceAllString(s, "N")
	s = strings.ReplaceAll(s, formatVerbSentinel, "%")
	return s
}

// firstLine extracts the first line of a description, trimmed and ending at
// the first period if present.
func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "\n"); idx != -1 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)
	if m := reFirstSentence.FindStringSubmatch(s); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if m := rePeriodAtEnd.FindStringSubmatch(s); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return s
}
