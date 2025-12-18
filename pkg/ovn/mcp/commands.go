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

// dangerousCharsPattern matches shell metacharacters that could enable command injection.
// We explicitly block: semicolons, ampersands, pipes, backticks, dollar signs, redirects, and backslashes.
var dangerousCharsPattern = regexp.MustCompile(`[;&|$` + "`" + `<>\\]`)

// dangerousCharsNoAmpPattern is like dangerousCharsPattern but allows & for microflow specs
// which use && for logical AND operations.
var dangerousCharsNoAmpPattern = regexp.MustCompile(`[;|$` + "`" + `<>\\]`)

// validateSafeString validates that a string doesn't contain dangerous shell metacharacters.
func validateSafeString(value, fieldName string, allowEmpty bool) error {
	if value == "" {
		if allowEmpty {
			return nil
		}
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	if dangerousCharsPattern.MatchString(value) {
		return fmt.Errorf("invalid %s: contains potentially dangerous characters", fieldName)
	}
	return nil
}

// validateRecordName validates that a record identifier is safe and non-empty.
func validateRecordName(record string) error {
	return validateSafeString(record, "record identifier", false)
}

// validateDatapath validates that a datapath name is safe and non-empty.
func validateDatapath(datapath string) error {
	return validateSafeString(datapath, "datapath name", false)
}

// validateMicroflow validates that a microflow specification is safe and non-empty.
// Microflow specs can contain && for logical AND, so we allow & but block other dangerous chars.
func validateMicroflow(microflow string) error {
	if microflow == "" {
		return fmt.Errorf("microflow specification cannot be empty")
	}

	if dangerousCharsNoAmpPattern.MatchString(microflow) {
		return fmt.Errorf("invalid microflow specification: contains potentially dangerous characters")
	}

	return nil
}

// validateColumnSpec validates that a column specification is safe (can be empty).
func validateColumnSpec(columns string) error {
	return validateSafeString(columns, "column specification", true)
}

// validateSwitchName validates that a logical switch name is safe (can be empty).
func validateSwitchName(switchName string) error {
	return validateSafeString(switchName, "switch name", true)
}

// validateEntityName validates that a logical switch or port group name is safe (can be empty).
func validateEntityName(entityName string) error {
	return validateSafeString(entityName, "entity name", true)
}

// validateRouterName validates that a logical router name is safe (can be empty).
func validateRouterName(routerName string) error {
	return validateSafeString(routerName, "router name", true)
}

// getDBCommand returns the appropriate command (ovn-nbctl or ovn-sbctl) for the given database.
func getDBCommand(db ovntypes.Database) string {
	if db == ovntypes.SouthboundDB {
		return "ovn-sbctl"
	}
	return "ovn-nbctl"
}
