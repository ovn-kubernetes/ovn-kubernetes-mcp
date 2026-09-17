// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package sosreport

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/headtail"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/pattern"
)

// getPodLogs reads container logs for a specific pod from the sosreport manifest.
// Log files use the containers naming layout:
// <pod_name>_<namespace>_<container_name>-<container_id>.log
// Returns the first matching container log (optionally filtered by container name).
func getPodLogs(sosreportPath, namespace, name, container string, patternParams pattern.PatternParams, headTailParams headtail.HeadTailParams) (string, error) {
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}
	if name == "" {
		return "", fmt.Errorf("name is required")
	}

	manifest, err := loadManifest(sosreportPath)
	if err != nil {
		return "", err
	}

	containerLogPlugin, exists := manifest.Components.Report.Plugins["container_log"]
	if !exists || len(containerLogPlugin.Files) == 0 {
		return "No pod logs found in sosreport\n", nil
	}

	var result []string
	matchedAnyFile := false

	for _, f := range containerLogPlugin.Files {
		for _, logPath := range f.FilesCopied {
			if !matchesContainerLogFile(logPath, namespace, name, container) {
				continue
			}
			matchedAnyFile = true

			// Remove the prefix if exists
			trimmedPath := strings.TrimPrefix(logPath, "host/")
			fullPath := filepath.Join(sosreportPath, trimmedPath)

			matches, err := searchInFile(fullPath, patternParams)
			if err != nil {
				return "", err
			}

			if len(matches) > 0 {
				result = append(result, matches...)
			}

			// Stop after the first matching container log.
			break
		}
		if matchedAnyFile {
			break
		}
	}

	if !matchedAnyFile {
		if container != "" {
			return fmt.Sprintf("No pod logs found for pod %s/%s container %q\n", namespace, name, container), nil
		}
		return fmt.Sprintf("No pod logs found for pod %s/%s\n", namespace, name), nil
	}
	if len(result) == 0 {
		if patternParams.Pattern != "" {
			return fmt.Sprintf("No matches found for pattern: %s\n", patternParams.Pattern), nil
		}
		return fmt.Sprintf("No log content found for pod %s/%s\n", namespace, name), nil
	}

	result = headTailParams.Apply(result, DefaultMaxLines)
	return strings.Join(result, "\n"), nil
}

// matchesContainerLogFile reports whether logPath is a container log for the
// given pod. Filenames follow:
// <pod_name>_<namespace>_<container_name>-<container_id>.log[.gz]
// When container is non-empty, the container_name must match exactly.
func matchesContainerLogFile(logPath, namespace, name, container string) bool {
	base := filepath.Base(logPath)
	base = strings.TrimSuffix(base, ".gz")
	if !strings.HasSuffix(base, ".log") {
		return false
	}
	base = strings.TrimSuffix(base, ".log")

	prefix := name + "_" + namespace + "_"
	if !strings.HasPrefix(base, prefix) {
		return false
	}
	remainder := strings.TrimPrefix(base, prefix)
	containerName, ok := parseContainerName(remainder)
	if !ok {
		return false
	}
	if container == "" {
		return true
	}
	return containerName == container
}

// parseContainerName extracts <container_name> from remainder formatted as
// <container_name>-<container_id>, where container_id is a non-empty hex string.
func parseContainerName(remainder string) (string, bool) {
	idx := strings.LastIndex(remainder, "-")
	if idx <= 0 {
		return "", false
	}
	containerName := remainder[:idx]
	idPart := remainder[idx+1:]
	if containerName == "" || idPart == "" {
		return "", false
	}
	for _, r := range idPart {
		if !unicode.Is(unicode.ASCII_Hex_Digit, r) {
			return "", false
		}
	}
	return containerName, true
}

// searchInFile searches in a file (handles both regular and gzip compressed files)
// Returns matching lines after pattern filtering
func searchInFile(filePath string, patternParams pattern.PatternParams) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("failed to close file %s: %v", filePath, err)
		}
	}()

	// Check if file is gzipped
	var reader io.Reader = file
	if strings.HasSuffix(filePath, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer func() {
			err := gzReader.Close()
			if err != nil {
				log.Printf("failed to close gzip reader %s: %v", filePath, err)
			}
		}()
		reader = gzReader
	}

	return patternParams.ExecuteWithMatch(func() ([]string, error) {
		lines, err := readLines(reader)
		if err != nil {
			return nil, err
		}
		return utils.StripEmptyLines(lines), nil
	}, true)
}
