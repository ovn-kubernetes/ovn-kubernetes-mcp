// SPDX-FileCopyrightText: Copyright The OVN-Kubernetes-MCP Contributors
// SPDX-License-Identifier: Apache-2.0

package sosreport

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/sosreport/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/headtail"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils/pattern"
)

const (
	// DefaultMaxLines is the default max lines returned when head/tail are unset.
	DefaultMaxLines = 100
	// DefaultMaxResults is the default max matches for sos-search-commands.
	DefaultMaxResults = 100
)

// getCommandOutput reads a command output file by filepath from manifest
func getCommandOutput(sosreportPath, relativeFilepath string, patternParams pattern.PatternParams, headTailParams headtail.HeadTailParams) (string, error) {
	if err := validateSosreportPath(sosreportPath); err != nil {
		return "", err
	}

	if err := validateRelativePath(relativeFilepath); err != nil {
		return "", err
	}

	fullPath := filepath.Join(sosreportPath, relativeFilepath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("command output file not found: %s", relativeFilepath)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("failed to close file %s: %v", file.Name(), err)
		}
	}()

	lines, err := patternParams.ExecuteWithMatch(func() ([]string, error) {
		fileLines, err := readLines(file)
		if err != nil {
			return nil, err
		}
		return utils.StripEmptyLines(fileLines), nil
	}, true)
	if err != nil {
		return "", err
	}

	if len(lines) == 0 && patternParams.Pattern != "" {
		return fmt.Sprintf("No lines matching pattern %q found\n", patternParams.Pattern), nil
	}

	lines = headTailParams.Apply(lines, DefaultMaxLines)
	return strings.Join(lines, "\n"), nil
}

// listPlugins returns a list of enabled plugins with their command counts
func listPlugins(sosreportPath string) (types.ListPluginsResult, error) {
	manifest, err := loadManifest(sosreportPath)
	if err != nil {
		return types.ListPluginsResult{}, err
	}

	var result types.ListPluginsResult
	totalCommands := 0

	// Only show enabled plugins
	for pluginName, plugin := range manifest.Components.Report.Plugins {
		commandCount := len(plugin.Commands)
		totalCommands += commandCount

		result.Plugins = append(result.Plugins, types.PluginSummary{
			Name:         pluginName,
			CommandCount: commandCount,
		})
	}

	result.TotalCommands = totalCommands
	return result, nil
}

// listCommands returns all commands for a specific plugin
func listCommands(sosreportPath, pluginName string) (types.ListCommandsResult, error) {
	manifest, err := loadManifest(sosreportPath)
	if err != nil {
		return types.ListCommandsResult{}, err
	}

	plugin, exists := manifest.Components.Report.Plugins[pluginName]
	if !exists {
		return types.ListCommandsResult{}, fmt.Errorf("plugin %q not found in manifest", pluginName)
	}

	result := types.ListCommandsResult{
		Plugin:       pluginName,
		CommandCount: len(plugin.Commands),
	}

	for _, cmd := range plugin.Commands {
		result.Commands = append(result.Commands, types.CommandSummary{
			Exec:     cmd.Exec,
			Filepath: cmd.Filepath,
		})
	}

	return result, nil
}

// searchCommands searches for commands matching a pattern across all plugins
func searchCommands(sosreportPath string, patternParams pattern.PatternParams, maxResults int) (types.SearchCommandsResult, error) {
	manifest, err := loadManifest(sosreportPath)
	if err != nil {
		return types.SearchCommandsResult{}, err
	}

	result := types.SearchCommandsResult{
		Matches: []types.CommandMatch{},
	}
	if maxResults <= 0 {
		maxResults = DefaultMaxResults
	}

	for pluginName, plugin := range manifest.Components.Report.Plugins {
		for _, cmd := range plugin.Commands {
			execMatches, err := patternParams.ExecuteWithMatch(func() ([]string, error) {
				return []string{cmd.Exec}, nil
			}, true)
			if err != nil {
				return types.SearchCommandsResult{}, err
			}

			filepathMatches, err := patternParams.ExecuteWithMatch(func() ([]string, error) {
				return []string{cmd.Filepath}, nil
			}, true)
			if err != nil {
				return types.SearchCommandsResult{}, err
			}

			if len(execMatches) == 1 || len(filepathMatches) == 1 {
				result.Matches = append(result.Matches, types.CommandMatch{
					Plugin:   pluginName,
					Exec:     cmd.Exec,
					Filepath: cmd.Filepath,
				})

				if len(result.Matches) >= maxResults {
					result.Total = len(result.Matches)
					return result, nil
				}
			}
		}
	}

	result.Total = len(result.Matches)
	return result, nil
}
