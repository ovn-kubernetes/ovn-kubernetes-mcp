package ovsdbtool

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// QueryDatabase queries the database for the given database name, table, where, and columns.
// It will return the result of the query.
func (s *OvsdbTool) QueryDatabase(ctx context.Context, mustGatherPath string, dbName string, table string, conditions []string, columns []string) (string, error) {
	// Validate the database name
	if dbName == "" || (!strings.HasSuffix(dbName, "_nbdb") && !strings.HasSuffix(dbName, "_sbdb")) {
		return "", fmt.Errorf("database name is required and must end with _nbdb or _sbdb: %s", dbName)
	}
	// Validate if the dbName does not contain any path traversal characters
	if filepath.Base(dbName) != dbName {
		return "", fmt.Errorf("database name %s contains path traversal characters", dbName)
	}

	// Validate the table
	if err := validateOVNTableName(table); err != nil {
		return "", fmt.Errorf("failed to validate table: %w", err)
	}

	// Get the schema name
	var schemaName string
	if strings.HasSuffix(dbName, "_nbdb") {
		schemaName = "OVN_Northbound"
	} else {
		schemaName = "OVN_Southbound"
	}

	// Build the query string
	queryString, err := buildQueryString(schemaName, table, conditions, columns)
	if err != nil {
		return "", fmt.Errorf("failed to build query string: %w", err)
	}

	// Get the path to the network logs directory
	networkLogsPath, err := getNetworkLogsDirectory(mustGatherPath)
	if err != nil {
		return "", fmt.Errorf("failed to get network logs directory: %w", err)
	}

	// Get the path to the database
	dbPath := filepath.Join(networkLogsPath, ovnkDatabaseStore, dbName)

	// Validate the database
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("database file %s not found: %w", dbPath, err)
		}
		return "", fmt.Errorf("failed to stat database file %s: %w", dbPath, err)
	}

	// Query the database
	output, err := exec.CommandContext(ctx, s.commandPath, "query", dbPath, queryString).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to query database. output: %s, error: %w", string(output), err)
	}

	return string(output), nil
}
