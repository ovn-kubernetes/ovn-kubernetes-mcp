package inspector

import (
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
	methodType   string
	toolName     string
	toolArgs     map[string]string
	commandflags map[string]string
}

func NewMCPInspector() *MCPInspector {
	return &MCPInspector{}
}

func (i *MCPInspector) Command(cmd string) *MCPInspector {
	i.command = cmd
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

func (i *MCPInspector) MethodCall(toolName string, toolArgs map[string]string) *MCPInspector {
	i.methodType = string(MethodTypeCall)
	i.toolName = toolName
	i.toolArgs = toolArgs
	return i
}

func (i *MCPInspector) Execute() ([]byte, error) {
	if i.command == "" {
		return nil, fmt.Errorf("command is required")
	}
	if i.methodType == "" {
		return nil, fmt.Errorf("method is required")
	}
	if i.methodType == string(MethodTypeCall) && i.toolName == "" {
		return nil, fmt.Errorf("tool name is required")
	}

	cmd, args, err := i.getCmdArgs()
	if err != nil {
		return nil, err
	}

	return exec.Command(cmd, args...).CombinedOutput()
}

func (i *MCPInspector) getCmdArgs() (string, []string, error) {
	cmd := "npx"
	args := []string{
		"-y",
		"@modelcontextprotocol/inspector",
		"--cli",
	}
	args = append(args, i.command)
	args = append(args, "--method")
	args = append(args, i.methodType)
	if i.methodType == string(MethodTypeCall) {
		args = append(args, "--tool-name")
		args = append(args, i.toolName)
		for key, value := range i.toolArgs {
			if value != "" {
				args = append(args, "--tool-arg")
				args = append(args, fmt.Sprintf("%s=%s", key, value))
			}
		}
	}

	if len(i.commandflags) > 0 {
		args = append(args, "--")
		for key, value := range i.commandflags {
			args = append(args, fmt.Sprintf("--%s", key))
			args = append(args, value)
		}
	}
	return cmd, args, nil
}
