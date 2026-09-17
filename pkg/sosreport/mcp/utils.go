// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package sosreport

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// validateSosreportPath validates that the path looks like a sosreport directory
func validateSosreportPath(sosreportPath string) error {
	if _, err := os.Stat(sosreportPath); os.IsNotExist(err) {
		return fmt.Errorf("sosreport path does not exist: %s", sosreportPath)
	}

	sosCommandsPath := filepath.Join(sosreportPath, "sos_commands")
	if _, err := os.Stat(sosCommandsPath); os.IsNotExist(err) {
		return fmt.Errorf("not a valid sosreport: missing sos_commands directory")
	}

	manifestPath := filepath.Join(sosreportPath, "sos_reports", "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("not a valid sosreport: missing sos_reports/manifest.json")
	}

	return nil
}

// validateRelativePath validates that a relative path doesn't attempt directory traversal
func validateRelativePath(relPath string) error {
	cleanPath := filepath.Clean(relPath)

	if strings.Contains(cleanPath, "..") || filepath.IsAbs(cleanPath) {
		return errors.New("invalid path: path traversal not allowed")
	}
	return nil
}

// readLines reads all lines from a reader using a large scanner buffer for long sos lines.
func readLines(reader io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(reader)

	// Increase buffer size for long lines
	// the size of initial allocation for buffer 4k
	buf := make([]byte, 4*1024)
	// the maximum size - 1M
	scanner.Buffer(buf, 1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
