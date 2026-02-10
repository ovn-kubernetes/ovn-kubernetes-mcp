//go:build ignore

// gen-readme-tools parses MCP tool definitions from pkg/*/mcp/mcp.go and
// updates the "Tools available in MCP Server" section in README.md.
// Live vs offline mode is inferred from cmd/ovnk-mcp-server/main.go.
package main

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
	mainPath   = "cmd/ovnk-mcp-server/main.go"
	readmeName = "README.md"
	// Sentinel for literal %% in format strings (multi-rune to avoid NUL and collisions with real text).
	formatVerbSentinel = "\uE000\uE001"
)

var (
	reFirstSentence = regexp.MustCompile(`^(.*?\.\s)`)
	rePeriodAtEnd   = regexp.MustCompile(`^(.*\.)$`)
	reFormatVerb    = regexp.MustCompile(`%[0-9.*+# \-]*[a-zA-Z]`)
)

type tool struct {
	name        string
	description string
}

func main() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "find repo root: %v\n", err)
		os.Exit(1)
	}
	pkgDir := filepath.Join(repoRoot, "pkg")
	mainFile := filepath.Join(repoRoot, mainPath)
	readmePath := filepath.Join(repoRoot, readmeName)

	liveOrder, offlineOrder, err := inferModesFromMain(mainFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "infer modes from main: %v\n", err)
		os.Exit(1)
	}

	liveTools := make(map[string][]tool)
	offlineTools := make(map[string][]tool)

	for _, pkgName := range liveOrder {
		mcpPath := filepath.Join(pkgDir, pkgName, "mcp", "mcp.go")
		if _, err := os.Stat(mcpPath); os.IsNotExist(err) {
			continue
		}
		tools, err := extractTools(mcpPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", mcpPath, err)
			continue
		}
		liveTools[pkgName] = tools
	}
	for _, pkgName := range offlineOrder {
		mcpPath := filepath.Join(pkgDir, pkgName, "mcp", "mcp.go")
		if _, err := os.Stat(mcpPath); os.IsNotExist(err) {
			continue
		}
		tools, err := extractTools(mcpPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", mcpPath, err)
			continue
		}
		offlineTools[pkgName] = tools
	}

	var out strings.Builder
	fmt.Fprintln(&out, "## Tools available in MCP Server")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "### Live Cluster Mode")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "Available when running with `--mode live-cluster` or `--mode dual` (and with a valid kubeconfig).")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "| Category      | Tool | Description |")
	fmt.Fprintln(&out, "|---------------|------|-------------|")
	for _, pkgName := range liveOrder {
		for i, t := range liveTools[pkgName] {
			prefix := "| **" + pkgName + "** |"
			if i > 0 {
				prefix = "| |"
			}
			fmt.Fprintf(&out, "%s `%s` | %s |\n", prefix, t.name, escapeTableCell(t.description))
		}
	}
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "### Offline Mode")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "Available when running with `--mode offline` or `--mode dual`. No cluster access required.")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "| Category       | Tool | Description |")
	fmt.Fprintln(&out, "|----------------|------|-------------|")
	for _, pkgName := range offlineOrder {
		for i, t := range offlineTools[pkgName] {
			prefix := "| **" + pkgName + "** |"
			if i > 0 {
				prefix = "| |"
			}
			fmt.Fprintf(&out, "%s `%s` | %s |\n", prefix, t.name, escapeTableCell(t.description))
		}
	}
	generated := out.String()
	if err := updateREADME(readmePath, generated); err != nil {
		fmt.Fprintf(os.Stderr, "update README: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Updated %s\n", readmePath)
}

// findRepoRoot returns the directory that contains go.mod.
func findRepoRoot() (string, error) {
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

// inferModesFromMain parses main.go and returns package names in order of first use in setupLiveCluster and setupOffline.
func inferModesFromMain(mainPath string) (liveOrder, offlineOrder []string, err error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, mainPath, nil, 0)
	if err != nil {
		return nil, nil, err
	}
	// Map import alias (e.g. kubernetesmcp or, when no explicit alias, "mcp") to pkg name (e.g. kubernetes) for .../pkg/<name>/mcp.
	importAliasToPkg := make(map[string]string)
	trim := "ovn-kubernetes-mcp/pkg/"
	for _, imp := range node.Imports {
		path := unquoteString(imp.Path.Value)
		idx := strings.Index(path, trim)
		if idx == -1 {
			continue
		}
		rest := path[idx+len(trim):]
		parts := strings.SplitN(rest, "/", 2)
		if len(parts) < 1 || parts[0] == "" {
			continue
		}
		pkgName := parts[0]
		alias := pkgName
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			// Default alias is the last segment of the import path (e.g. "mcp" for .../pkg/kubernetes/mcp).
			alias = filepath.Base(path)
		}
		// Warn if the alias maps to multiple packages. This is unlikely to happen
		// in practice, but it's possible if the import path is ambiguous.
		if existing, ok := importAliasToPkg[alias]; ok && existing != pkgName {
			fmt.Fprintf(os.Stderr, "warning: import alias %q maps to dual %q and %q; results may be incorrect\n", alias, existing, pkgName)
		}
		importAliasToPkg[alias] = pkgName
	}

	var setupLiveClusterBody *ast.BlockStmt
	var setupOfflineBody *ast.BlockStmt
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		switch fn.Name.Name {
		case "setupLiveCluster":
			setupLiveClusterBody = fn.Body
		case "setupOffline":
			setupOfflineBody = fn.Body
		}
	}
	if setupLiveClusterBody == nil || setupOfflineBody == nil {
		return nil, nil, fmt.Errorf("setupLiveCluster or setupOffline not found in %s", mainPath)
	}

	liveOrder = collectPackagesInOrder(setupLiveClusterBody, importAliasToPkg)
	offlineOrder = collectPackagesInOrder(setupOfflineBody, importAliasToPkg)
	return liveOrder, offlineOrder, nil
}

// collectPackagesInOrder walks the block and returns pkg names in order of first occurrence.
func collectPackagesInOrder(block *ast.BlockStmt, aliasToPkg map[string]string) []string {
	seen := make(map[string]bool)
	var order []string
	ast.Inspect(block, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if pkg, ok := aliasToPkg[ident.Name]; ok && !seen[pkg] {
			seen[pkg] = true
			order = append(order, pkg)
		}
		return true
	})
	return order
}

// firstLine extracts the first line of a description, trimmed and ending at first period if present.
func firstLine(s string) string {
	s = strings.TrimSpace(s)
	// Take first line (up to newline)
	if idx := strings.Index(s, "\n"); idx != -1 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)
	// Take first sentence: stop at ". " or "." at end-of-string
	if m := reFirstSentence.FindStringSubmatch(s); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	// If no ". " found, try period at end of string
	if m := rePeriodAtEnd.FindStringSubmatch(s); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return s
}

// extractTools parses the given mcpPath file and returns all MCP tools registered
// via mcp.AddTool(server, &mcp.Tool{...}). Each tool's description is shortened to
// the first line or first sentence for the README table. Each tool's short description
// should be provided in the first line or sentence for this function to work correctly.
func extractTools(mcpPath string) ([]tool, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, mcpPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var tools []tool
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}
		// Check for mcp.AddTool(server, &mcp.Tool{...})
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "AddTool" {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok || ident.Name != "mcp" {
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
					name = unquoteString(bl.Value)
				}
			case "Description":
				desc = extractDescription(kv.Value)
			}
		}
		if name != "" {
			tools = append(tools, tool{name: name, description: firstLine(desc)})
		}
		return true
	})
	return tools, nil
}

// extractDescription returns the string value of a Tool's Description field from the AST.
// It handles dual plain string literals and fmt.Sprintf(format, ...) calls; format verbs
// in the latter are stripped.
func extractDescription(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.BasicLit:
		if v.Kind == token.STRING {
			return unquoteString(v.Value)
		}
	case *ast.CallExpr:
		// fmt.Sprintf(`...`, ...) - first argument is format string; replace format verbs so they don't appear literally in README
		if len(v.Args) >= 1 {
			if bl, ok := v.Args[0].(*ast.BasicLit); ok && bl.Kind == token.STRING {
				return stripFormatVerbs(unquoteString(bl.Value))
			}
		}
	}
	return ""
}

// stripFormatVerbs replaces Go format verbs (e.g. %d, %s) with placeholders so format strings
// from fmt.Sprintf read naturally in the README.
func stripFormatVerbs(s string) string {
	// Replace %% with a sentinel so we don't touch literal percents (multi-rune to avoid NUL and collisions)
	s = strings.ReplaceAll(s, "%%", formatVerbSentinel)
	s = reFormatVerb.ReplaceAllString(s, "N")
	s = strings.ReplaceAll(s, formatVerbSentinel, "%")
	return s
}

// escapeTableCell escapes content for use in a markdown table cell so that | and newlines don't
// break the table.
func escapeTableCell(s string) string {
	s = strings.ReplaceAll(s, "|", "&#124;")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

// unquoteString removes surrounding backticks or double quotes from a Go string literal.
// It uses strconv.Unquote to handle escape sequences. Returns the unquoted string or the
// original string if unquoting fails.
func unquoteString(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '`' && s[len(s)-1] == '`' {
		return s[1 : len(s)-1]
	}
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"') {
		// Try to unquote using strconv.Unquote for escape sequences. If it fails,
		// use simple unquote.
		if u, err := strconv.Unquote(s); err == nil {
			return u
		}
		return s[1 : len(s)-1]
	}
	return s
}

// readREADME reads and returns the full contents of the file at path.
func readREADME(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// writeREADME writes content to the file at path, creating or truncating it.
func writeREADME(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// updateREADME replaces the "Tools available" section in README with generated content.
func updateREADME(readmePath, generated string) error {
	content, err := readREADME(readmePath)
	if err != nil {
		return err
	}
	startMark := "<!-- TOOLS_SECTION_START -->"
	endMark := "<!-- TOOLS_SECTION_END -->"
	startIdx := strings.Index(content, startMark)
	endIdx := strings.Index(content, endMark)
	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return fmt.Errorf("README must contain %s and %s", startMark, endMark)
	}
	// Before: everything up to and including the newline after startMark
	afterStart := startIdx + len(startMark)
	if afterStart < len(content) && content[afterStart] == '\n' {
		afterStart++
	}
	before := content[:afterStart]
	after := content[endIdx:]
	newContent := before + generated + "\n" + after
	return writeREADME(readmePath, newContent)
}
