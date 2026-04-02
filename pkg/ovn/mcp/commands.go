package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	k8stypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
	ovntypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/ovn/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
)

const defaultMaxLines = 100

func (s *MCPServer) runCommand(ctx context.Context, req *mcp.CallToolRequest, namespacedName k8stypes.NamespacedNameParams,
	commands []string) ([]string, error) {
	_, result, err := s.k8sMcpServer.ExecPod(ctx, req, k8stypes.ExecPodParams{NamespacedNameParams: namespacedName, Command: commands})
	if err != nil {
		return nil, err
	}
	if result.Stderr != "" {
		return nil, fmt.Errorf("error occurred while running command %v on pod %s/%s: %s", commands, namespacedName.Namespace,
			namespacedName.Name, result.Stderr)
	}
	return utils.StripEmptyLines(strings.Split(result.Stdout, "\n")), nil
}

// validateDatabase validates that the database is a valid OVN database.
func validateDatabase(db ovntypes.Database) error {
	switch db {
	case ovntypes.NorthboundDB, ovntypes.SouthboundDB:
		return nil
	default:
		return fmt.Errorf("invalid database %q: must be 'nbdb' or 'sbdb'", db)
	}
}

// validateRecordName validates that a record identifier is safe and non-empty.
func validateRecordName(record string) error {
	return utils.ValidateSafeString(record, "record identifier", false, utils.ShellMetaCharactersTypeDefault)
}

// validateDatapath validates that a datapath name is safe and non-empty.
func validateDatapath(datapath string) error {
	return utils.ValidateSafeString(datapath, "datapath name", false, utils.ShellMetaCharactersTypeDefault)
}

// validateMicroflow validates that a microflow specification is safe and non-empty.
// Microflow specs can contain && for logical AND, so we allow & but block other dangerous chars.
func validateMicroflow(microflow string) error {
	return utils.ValidateSafeString(microflow, "microflow specification", false, utils.ShellMetaCharactersTypeAllowAmp)
}

// validateColumnSpec validates that a column specification is safe (can be empty).
func validateColumnSpec(columns string) error {
	return utils.ValidateSafeString(columns, "column specification", true, utils.ShellMetaCharactersTypeDefault)
}

// getDBCommand returns the appropriate command (ovn-nbctl or ovn-sbctl) for the given database.
func getDBCommand(db ovntypes.Database) string {
	if db == ovntypes.SouthboundDB {
		return "ovn-sbctl"
	}
	return "ovn-nbctl"
}
