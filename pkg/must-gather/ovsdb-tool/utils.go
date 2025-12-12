package ovsdbtool

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	mustgatherUtils "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/utils"
)

// getOvsdbToolCommandPath returns the path to the ovsdb-tool command
func getOvsdbToolCommandPath() (string, error) {
	// Check if ovsdb-tool is in the PATH
	path, err := exec.LookPath("ovsdb-tool")
	if err != nil {
		return "", fmt.Errorf("failed to look path for ovsdb-tool: %w", err)
	}
	return path, nil
}

// getNetworkLogsDirectory gets the path to the network logs directory in the must gather path.
// It will return the path to the network logs directory. It will return an error if the
// must gather path is not valid or contains multiple directories or the network logs directory
// does not exist.
func getNetworkLogsDirectory(mustgatherPath string) (string, error) {
	// Validate the must gather path
	err := mustgatherUtils.ValidateMustGatherPath(mustgatherPath)
	if err != nil {
		return "", fmt.Errorf("failed to validate must gather path: %w", err)
	}

	// Get the name of the only directory in the must gather path
	dirs, err := os.ReadDir(mustgatherPath)
	if err != nil {
		return "", fmt.Errorf("failed to read must gather path: %w", err)
	}
	var dirName string
	count := 0
	for _, dir := range dirs {
		if dir.IsDir() {
			dirName = dir.Name()
			count++
			if count > 1 {
				return "", fmt.Errorf("must gather path contains multiple directories: %v", dirs)
			}
		}
	}
	if count == 0 {
		return "", fmt.Errorf("must gather path does not contain any directories: %v", dirs)
	}

	networkLogsPath := filepath.Join(mustgatherPath, dirName, networkLogsDirectory)

	// Check if the network logs path exists
	if _, err := os.Stat(networkLogsPath); os.IsNotExist(err) {
		return "", fmt.Errorf("network logs path does not exist: %w", err)
	}

	return networkLogsPath, nil
}

// isWritable checks if the given path is writable. It will return true if the path is writable,
// false otherwise. It will return an error if the path is not valid.
func isWritable(path string) (bool, error) {
	tmpFile := "tmpfile"

	file, err := os.CreateTemp(path, tmpFile)
	if err != nil {
		return false, fmt.Errorf("failed to create temporary file: %w", err)
	}

	defer os.Remove(file.Name())
	defer file.Close()

	return true, nil
}

// extractOvnDatabases extracts the ovn databases from the must gather path.
// It will return the path to the extracted dbs. It will return an error if
// the dbs cannot be extracted.
func extractOvnDatabases(mustgatherPath string) (string, error) {
	// Get the path to the network logs directory
	networkLogsPath, err := getNetworkLogsDirectory(mustgatherPath)
	if err != nil {
		return "", fmt.Errorf("failed to get network logs directory: %w", err)
	}

	// Check if write permissions are available in the network logs path
	writable, err := isWritable(networkLogsPath)
	if err != nil {
		return "", fmt.Errorf("failed to check if network logs path is writable: %w", err)
	}
	if !writable {
		return "", fmt.Errorf("network logs path does not have write permissions")
	}

	// Check if the dbs are already extracted
	dbPath := filepath.Join(networkLogsPath, ovnkDatabaseStore)
	stat, err := os.Stat(dbPath)
	if err == nil && stat.IsDir() {
		return dbPath, nil
	}

	// If the db path is a file, remove it
	if err == nil && !stat.IsDir() {
		err = os.RemoveAll(dbPath)
		if err != nil {
			return "", fmt.Errorf("failed to remove db directory: %w", err)
		}
	}

	// Check if the tar file exists
	dbTarPath := filepath.Join(networkLogsPath, fmt.Sprintf("%s.tar.gz", ovnkDatabaseStore))
	if _, err := os.Stat(dbTarPath); os.IsNotExist(err) {
		return "", fmt.Errorf("ovnk database store tar file not found: %w", err)
	}

	// Extract the dbs
	gzFile, err := os.Open(dbTarPath)
	if err != nil {
		return "", fmt.Errorf("failed to open gz file: %w", err)
	}
	defer gzFile.Close()

	gzReader, err := gzip.NewReader(gzFile)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}
		// Create the directory if the type flag is a directory
		if header.Typeflag == tar.TypeDir {
			destPath := filepath.Join(networkLogsPath, header.Name)
			if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(networkLogsPath)) {
				return "", fmt.Errorf("invalid directory path in archive: %s", header.Name)
			}
			err = os.MkdirAll(destPath, 0755)
			if err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			// Create the file if the type flag is not a directory
			err := func() error {
				destPath := filepath.Join(networkLogsPath, header.Name)
				if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(networkLogsPath)) {
					return fmt.Errorf("invalid file path in archive: %s", header.Name)
				}
				dbFile, err := os.Create(destPath)
				if err != nil {
					return fmt.Errorf("failed to create db file: %w", err)
				}
				defer dbFile.Close()
				_, err = io.Copy(dbFile, tarReader)
				if err != nil {
					return fmt.Errorf("failed to copy db file: %w", err)
				}
				return nil
			}()
			if err != nil {
				return "", fmt.Errorf("failed to create db file: %w", err)
			}
		}
	}

	return dbPath, nil
}

// buildQueryString builds the query string for the given schema name, table, conditions, and columns.
// It will return the query string. The value for schemaName and table should not be empty. The schema name
// should be either "OVN_Northbound" or "OVN_Southbound".
func buildQueryString(schemaName string, table string, conditions []string, columns []string) (string, error) {
	// Validate the schema name
	if schemaName != "OVN_Northbound" && schemaName != "OVN_Southbound" {
		return "", fmt.Errorf("schema name must be either \"OVN_Northbound\" or \"OVN_Southbound\": %s", schemaName)
	}

	// Validate the table
	if table == "" {
		return "", fmt.Errorf("table must not be empty")
	}

	var sb strings.Builder
	// Write the schema name and table
	sb.WriteString(fmt.Sprintf("[\"%s\", {\"op\": \"select\", \"table\": \"%s\", \"where\": [", schemaName, table))
	// Write the conditions
	for i, condition := range conditions {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(condition)
	}
	sb.WriteString("]")
	// Write the columns if there are any
	if len(columns) > 0 {
		sb.WriteString(", \"columns\": [")
		for i, column := range columns {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("\"%s\"", column))
		}
		sb.WriteString("]")
	}
	sb.WriteString("}]")

	return sb.String(), nil
}
