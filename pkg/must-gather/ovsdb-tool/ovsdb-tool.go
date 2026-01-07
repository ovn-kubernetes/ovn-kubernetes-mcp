package ovsdbtool

import (
	"fmt"

	omcclient "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/omc-client"
)

type OvsdbTool struct {
	commandPath string
	omcClient   *omcclient.OmcClient
}

func NewOvsdbTool(omcClient *omcclient.OmcClient) (*OvsdbTool, error) {
	commandPath, err := getOvsdbToolCommandPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get ovsdb tool command path: %w", err)
	}
	return &OvsdbTool{
		commandPath: commandPath,
		omcClient:   omcClient,
	}, nil
}
