package sosreport

import (
	"compress/gzip"
	"os"
	"strings"
	"path/filepath"
	"regexp"
	"fmt"
	"io"
)

// searchPodLogs searches pod logs using the manifest to find log files
func searchPodLogs(sosreportPath, pattern, podFilter string, maxResults int) (string, error) {
	searchPattern, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid search pattern: %w", err)
	}

	manifest, err := loadManifest(sosreportPath)
	if err != nil {
		return "", err
	}

	if maxResults <= 0 {
		maxResults = defaultResultLimit
	}

	containerLogPlugin, exists := manifest.Components.Report.Plugins["container_log"]
	if !exists || len(containerLogPlugin.Files) == 0 {
		return "No pod logs found in sosreport\n", nil
	}

	var result strings.Builder
	totalMatches := 0

	for _, f := range containerLogPlugin.Files {
		for _, logPath := range f.FilesCopied {
			if podFilter != "" && !strings.Contains(logPath, podFilter) {
				continue
			}

			// Remove the prefix if exists
			logPath = strings.TrimPrefix(logPath, "host/")

			fullPath := filepath.Join(sosreportPath, logPath)

			matches, err := searchInFile(fullPath, searchPattern, maxResults-totalMatches)
			if err != nil {
				return "", err
			}

			if len(matches) > 0 {
				result.WriteString(fmt.Sprintf("\n=== %s ===\n", logPath))
				result.WriteString(matches)
				totalMatches += len(strings.Split(matches, "\n"))

				if totalMatches >= maxResults {
					result.WriteString(fmt.Sprintf("\n... (search truncated at %d results)\n", maxResults))
					return result.String(), nil
				}
			}
		}
	}
	if totalMatches == 0 {
		return fmt.Sprintf("No matches found for pattern: %s\n", pattern), nil
	}

	return result.String(), nil
}

// searchInFile searches in a file (handles both regular and gzip compressed files)
// Returns the matches and an error
func searchInFile(filePath string, pattern *regexp.Regexp, maxLines int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Check if file is gzipped
	var reader io.Reader = file
	if strings.HasSuffix(filePath, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return "", fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Read with limit directly from the reader
	return readWithLimit(reader, pattern, maxLines)
}
