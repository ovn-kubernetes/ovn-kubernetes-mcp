package sosreport

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"regexp"
	"bufio"
	"errors"
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

// readWithLimit reads from a reader with a line limit
func readWithLimit(reader io.Reader, pattern *regexp.Regexp, maxLines int) (string, error) {
	var result strings.Builder
	scanner := bufio.NewScanner(reader)

	// Increase buffer size for long lines
	// the size of initial allocation for buffer 4k
	buf := make([]byte, 4*1024)
	// the maximum size - 1M
	scanner.Buffer(buf, 1024*1024)

	lineCount := 0
	for scanner.Scan() {
		if pattern == nil || pattern.MatchString(scanner.Text()) {
			result.WriteString(scanner.Text())
			result.WriteString("\n")
			lineCount++
		}

		if maxLines > 0 && lineCount >= maxLines {
			result.WriteString(fmt.Sprintf("\n... (output truncated at %d lines)\n", maxLines))
			break
		}

	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return result.String(), nil
}
