package types

// ListCommandsParams are the parameters for sos-list-commands
type ListCommandsParams struct {
	SosreportPath string `json:"sosreport_path"`
	Plugin        string `json:"plugin"`
}

// ListCommandsResult returns commands for a specific plugin
type ListCommandsResult struct {
	Plugin       string           `json:"plugin"`
	CommandCount int              `json:"command_count"`
	Commands     []CommandSummary `json:"commands"`
}

// CommandSummary provides a summary of a command
type CommandSummary struct {
	Exec     string `json:"exec"`
	Filepath string `json:"filepath"`
}

// SearchCommandsParams are the parameters for sos-search-commands
type SearchCommandsParams struct {
	SosreportPath string `json:"sosreport_path"`
	Pattern       string `json:"pattern"`
	MaxResults    int    `json:"max_results,omitempty"`
}

// SearchCommandsResult returns commands matching the pattern
type SearchCommandsResult struct {
	Matches []CommandMatch `json:"matches"`
	Total   int            `json:"total"`
}

// CommandMatch represents a command that matches the search
type CommandMatch struct {
	Plugin   string `json:"plugin"`
	Exec     string `json:"exec"`
	Filepath string `json:"filepath"`
}

// GetCommandParams are the parameters for sos-get-command
type GetCommandParams struct {
	SosreportPath string `json:"sosreport_path"`
	Filepath      string `json:"filepath"`
	Pattern       string `json:"pattern,omitempty"`
	MaxLines      int    `json:"max_lines,omitempty"`
}

// GetCommandResult returns the command output
type GetCommandResult struct {
	Output string `json:"output"`
}

// ListPluginsResult returns plugins with command counts
type ListPluginsResult struct {
	Plugins       []PluginSummary `json:"plugins"`
	TotalCommands int             `json:"total_commands"`
}

// ListPluginsParams are the parameters for sos-list-plugins
type ListPluginsParams struct {
	SosreportPath string `json:"sosreport_path"`
}

// PluginSummary provides a summary of a plugin
type PluginSummary struct {
	Name         string `json:"name"`
	CommandCount int    `json:"command_count"`
}
