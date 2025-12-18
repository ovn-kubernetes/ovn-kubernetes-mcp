package mcp

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	k8stypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	ovntypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/ovn/types"
)

const defaultMaxLines = 1000

func (s *MCPServer) runCommand(ctx context.Context, req *mcp.CallToolRequest, namespacedName k8stypes.NamespacedNameParams,
	commands []string) ([]string, error) {
	_, result, err := s.k8sMcpServer.ExecPod(ctx, req, k8stypes.ExecPodParams{NamespacedNameParams: namespacedName, Command: commands})
	if err != nil {
		return nil, err
	}
	if result.Stderr != "" {
		return nil, fmt.Errorf("error occurred while running command %v on pod %s/%s: %s", commands, namespacedName.Namespace,
			namespacedName.Name, result.Stderr)
	}
	return parseOutput(result.Stdout), nil
}

// parseOutput parses command output into lines, trimming whitespace and removing empty lines.
func parseOutput(stdout string) []string {
	output := []string{} // Initialize with empty slice to ensure valid JSON when there's no output
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			output = append(output, line)
		}
	}
	return output
}

// filterLines filters lines using a regex pattern.
func filterLines(lines []string, pattern string) ([]string, error) {
	if pattern == "" {
		return lines, nil
	}

	filterPattern, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid filter pattern %s: %w", pattern, err)
	}

	filtered := []string{} // Initialize with empty slice to ensure valid JSON when there's no output
	for _, line := range lines {
		if filterPattern.MatchString(line) {
			filtered = append(filtered, line)
		}
	}
	return filtered, nil
}

// limitLines limits the number of lines returned.
func limitLines(lines []string, maxLines int) []string {
	if maxLines <= 0 {
		maxLines = defaultMaxLines
	}
	if len(lines) > maxLines {
		return lines[:maxLines]
	}
	return lines
}

// validateDatabase validates that the database is a valid OVN database.
func validateDatabase(db ovntypes.Database) error {
	switch db {
	case ovntypes.NorthboundDB, ovntypes.SouthboundDB:
		return nil
	default:
		return fmt.Errorf("invalid database %q: must be 'nbdb' or 'sbdb'", db)
	}
}

// validateTableName validates that a table name is safe and non-empty.
// Table names should only contain alphanumeric characters and underscores.
func validateTableName(table string) error {
	if table == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// OVN table names follow naming conventions: alphanumeric and underscores
	validTableName := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	if !validTableName.MatchString(table) {
		return fmt.Errorf("invalid table name %q: must start with a letter and contain only alphanumeric characters and underscores", table)
	}

	return nil
}

// validateRecordName validates that a record identifier is safe and non-empty.
func validateRecordName(record string) error {
	if record == "" {
		return fmt.Errorf("record identifier cannot be empty")
	}

	// Check for potentially dangerous characters that shouldn't appear in record identifiers
	// Record identifiers can be UUIDs, names, or other identifiers
	// We explicitly block: semicolons, pipes, backticks, dollar signs, and other shell metacharacters
	dangerousChars := regexp.MustCompile(`[;&|$` + "`" + `<>\\]`)
	if dangerousChars.MatchString(record) {
		return fmt.Errorf("invalid record identifier: contains potentially dangerous characters")
	}

	return nil
}

// validateDatapath validates that a datapath name is safe and non-empty.
func validateDatapath(datapath string) error {
	if datapath == "" {
		return fmt.Errorf("datapath name cannot be empty")
	}

	// Check for potentially dangerous characters
	dangerousChars := regexp.MustCompile(`[;&|$` + "`" + `<>\\]`)
	if dangerousChars.MatchString(datapath) {
		return fmt.Errorf("invalid datapath name: contains potentially dangerous characters")
	}

	return nil
}

// validateMicroflow validates that a microflow specification is safe and non-empty.
func validateMicroflow(microflow string) error {
	if microflow == "" {
		return fmt.Errorf("microflow specification cannot be empty")
	}

	// Check for potentially dangerous characters that shouldn't appear in microflow specs
	// Microflow specs can contain && for logical AND, so we allow & but block single dangerous chars
	// We explicitly block: semicolons, pipes, backticks, dollar signs, and other shell metacharacters
	dangerousChars := regexp.MustCompile(`[;|$` + "`" + `<>\\]`)
	if dangerousChars.MatchString(microflow) {
		return fmt.Errorf("invalid microflow specification: contains potentially dangerous characters")
	}

	return nil
}

// validateColumnSpec validates that a column specification is safe (can be empty).
func validateColumnSpec(columns string) error {
	if columns == "" {
		return nil
	}

	// Check for potentially dangerous characters
	dangerousChars := regexp.MustCompile(`[;&|$` + "`" + `<>\\]`)
	if dangerousChars.MatchString(columns) {
		return fmt.Errorf("invalid column specification: contains potentially dangerous characters")
	}

	return nil
}

// validateSwitchName validates that a logical switch name is safe (can be empty).
func validateSwitchName(switchName string) error {
	if switchName == "" {
		return nil // switch name can be empty for listing all
	}

	// Check for potentially dangerous characters
	dangerousChars := regexp.MustCompile(`[;&|$` + "`" + `<>\\]`)
	if dangerousChars.MatchString(switchName) {
		return fmt.Errorf("invalid switch name: contains potentially dangerous characters")
	}

	return nil
}

// validateEntityName validates that a logical switch or port group name is safe (can be empty).
func validateEntityName(entityName string) error {
	if entityName == "" {
		return nil // entity name can be empty for listing all
	}

	// Check for potentially dangerous characters
	dangerousChars := regexp.MustCompile(`[;&|$` + "`" + `<>\\]`)
	if dangerousChars.MatchString(entityName) {
		return fmt.Errorf("invalid entity name: contains potentially dangerous characters")
	}

	return nil
}

// validateRouterName validates that a logical router name is safe (can be empty).
func validateRouterName(routerName string) error {
	if routerName == "" {
		return nil // router name can be empty for listing all
	}

	// Check for potentially dangerous characters
	dangerousChars := regexp.MustCompile(`[;&|$` + "`" + `<>\\]`)
	if dangerousChars.MatchString(routerName) {
		return fmt.Errorf("invalid router name: contains potentially dangerous characters")
	}

	return nil
}

// getDBCommand returns the appropriate command (ovn-nbctl or ovn-sbctl) for the given database.
func getDBCommand(db ovntypes.Database) string {
	if db == ovntypes.SouthboundDB {
		return "ovn-sbctl"
	}
	return "ovn-nbctl"
}
