package types

// Manifest represents the sosreport manifest.json structure
type Manifest struct {
	Version    string     `json:"version"`
	Cmdline    string     `json:"cmdline"`
	StartTime  string     `json:"start_time"`
	EndTime    string     `json:"end_time"`
	RunTime    string     `json:"run_time"`
	Components Components `json:"components"`
}

// Components contains the report components
type Components struct {
	Report Report `json:"report"`
}

// Report contains the sosreport details
type Report struct {
	Plugins map[string]PluginDetail `json:"plugins"`
}

// PluginDetail contains details about a plugin
type PluginDetail struct {
	StartTime   string          `json:"start_time"`
	EndTime     string          `json:"end_time"`
	RunTime     string          `json:"run_time"`
	Commands    []CommandDetail `json:"commands"`
	Files       []FilesDetail   `json:"files"`
	Containers  interface{}     `json:"containers"`
	Collections interface{}     `json:"collections"`
}

// CommandDetail contains details about a command execution
type CommandDetail struct {
	Command    string   `json:"command"`
	Parameters []string `json:"parameters"`
	Exec       string   `json:"exec"`
	Filepath   string   `json:"filepath"`
	Truncated  bool     `json:"truncated"`
	ReturnCode int      `json:"return_code"`
	Priority   int      `json:"priority"`
	StartTime  float64  `json:"start_time"`
	EndTime    float64  `json:"end_time"`
	RunTime    float64  `json:"run_time"`
	Tags       []string `json:"tags"`
}

// FilesDetail contains details about files collected
type FilesDetail struct {
	FilesCopied []string `json:"files_copied"`
	Tags        []string `json:"tags"`
}
