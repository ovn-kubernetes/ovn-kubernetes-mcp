package inspector

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type MethodType string

const (
	MethodTypeList MethodType = "tools/list"
	MethodTypeCall MethodType = "tools/call"
)

type MCPInspector struct {
	command      string
	serverURL    string // when set gets priority over command, connect via HTTP
	methodType   string
	toolName     string
	toolArgs     map[string]any
	commandflags map[string]string
}

func NewMCPInspector() *MCPInspector {
	return &MCPInspector{}
}

func (i *MCPInspector) Command(cmd string) *MCPInspector {
	i.command = cmd
	return i
}

// URL configures the inspector to connect to an MCP server over HTTP at the given URL
// (e.g. http://127.0.0.1:18080 when using port-forward to a deployed server).
func (i *MCPInspector) URL(url string) *MCPInspector {
	i.serverURL = url
	return i
}

func (i *MCPInspector) CommandFlags(env map[string]string) *MCPInspector {
	i.commandflags = env
	return i
}

func (i *MCPInspector) MethodList() *MCPInspector {
	i.methodType = string(MethodTypeList)
	return i
}

func (i *MCPInspector) MethodCall(toolName string, toolArgs map[string]any) *MCPInspector {
	i.methodType = string(MethodTypeCall)
	i.toolName = toolName
	i.toolArgs = toolArgs
	return i
}

func (i *MCPInspector) Execute() ([]byte, error) {
	if i.command == "" && i.serverURL == "" {
		return nil, fmt.Errorf("command or server URL is required")
	}
	if i.methodType == "" {
		return nil, fmt.Errorf("method is required")
	}
	if i.methodType == string(MethodTypeCall) && i.toolName == "" {
		return nil, fmt.Errorf("tool name is required")
	}

	cmdName, args, err := i.getCmdArgs()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(cmdName, args...)
	// Capture stdout and stderr separately
	// npx prints installation messages to stderr which would corrupt the JSON output
	output, err := cmd.Output()
	if err != nil {
		// Include stderr in error message for debugging
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("command failed: %v, stderr: %s", err, string(exitErr.Stderr))
		}
		return nil, err
	}
	return output, nil
}

func (i *MCPInspector) getCmdArgs() (string, []string, error) {
	cmd := "npx"
	args := []string{
		"-y",
		"@modelcontextprotocol/inspector",
		"--cli",
	}
	if i.serverURL != "" {
		args = append(args, i.serverURL)
		args = append(args, "--transport", "http")
	} else {
		args = append(args, i.command)
	}
	args = append(args, "--method")
	args = append(args, i.methodType)
	if i.methodType == string(MethodTypeCall) {
		args = append(args, "--tool-name")
		args = append(args, i.toolName)
		for key, value := range i.toolArgs {
			var valueStr string
			switch v := value.(type) {
			case string:
				valueStr = v
			case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				valueStr = fmt.Sprint(v)
			case []string:
				if len(v) == 0 {
					continue
				}
				b, err := json.Marshal(v)
				if err != nil {
					return "", nil, fmt.Errorf("serialize tool-arg %s: %w", key, err)
				}
				valueStr = string(b)
			default:
				return "", nil, fmt.Errorf("unsupported tool-arg type for %s: %T", key, value)
			}
			if valueStr != "" {
				args = append(args, "--tool-arg")
				args = append(args, fmt.Sprintf("%s=%s", key, valueStr))
			}
		}
	}

	// Command flags (e.g. --kubeconfig) only apply when running the server as a subprocess
	if i.serverURL == "" && len(i.commandflags) > 0 {
		args = append(args, "--")
		for key, value := range i.commandflags {
			args = append(args, fmt.Sprintf("--%s", key))
			args = append(args, value)
		}
	}
	return cmd, args, nil
}
