package sosreport

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/sosreport/types"
)

const (
	// defaultResultLimit is the default maximum number of lines/results to return
	defaultResultLimit = 100
)

// getCommandOutput reads a command output file by filepath from manifest
func getCommandOutput(sosreportPath, relativeFilepath, pattern string, maxLines int) (string, error) {
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
	defer file.Close()

	var searchPattern *regexp.Regexp
	if pattern != "" {
		searchPattern, err = regexp.Compile(pattern)
		if err != nil {
			return "", fmt.Errorf("invalid pattern: %w", err)
		}
	}

	if maxLines <= 0 {
		maxLines = defaultResultLimit
	}

	output, err := readWithLimit(file, searchPattern, maxLines)
	if err != nil {
		return "", err
	}

	if output == "" && pattern != "" {
		return fmt.Sprintf("No lines matching pattern %q found\n", pattern), nil
	}

	return output, nil
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
func searchCommands(sosreportPath, pattern string, maxResults int) (types.SearchCommandsResult, error) {
	manifest, err := loadManifest(sosreportPath)
	if err != nil {
		return types.SearchCommandsResult{}, err
	}

	searchPattern, err := regexp.Compile(pattern)
	if err != nil {
		return types.SearchCommandsResult{}, fmt.Errorf("invalid search pattern: %w", err)
	}

	var result types.SearchCommandsResult
	if maxResults <= 0 {
		maxResults = defaultResultLimit
	}

	for pluginName, plugin := range manifest.Components.Report.Plugins {
		for _, cmd := range plugin.Commands {
			if searchPattern.MatchString(cmd.Exec) || searchPattern.MatchString(cmd.Filepath) {
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
