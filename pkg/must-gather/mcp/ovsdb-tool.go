package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/types"
)

// ListNorthboundDatabases lists the Northbound Databases in the must gather path.
// It will return the Northbound Database to node mapping.
func (s *MustGatherMCPServer) ListNorthboundDatabases(ctx context.Context, req *mcp.CallToolRequest, in types.ListDatabasesParams) (*mcp.CallToolResult, types.ListDatabasesResult, error) {
	if s.ovsdbTool == nil {
		return nil, types.ListDatabasesResult{}, fmt.Errorf("ovsdb-tool is not available; ensure ovsdb-tool binary is in PATH")
	}
	output, err := s.ovsdbTool.ListNorthboundDatabases(ctx, in.MustGatherPath)
	if err != nil {
		return nil, types.ListDatabasesResult{}, err
	}
	return nil, types.ListDatabasesResult{Data: output}, nil
}

// ListSouthboundDatabases lists the Southbound Databases in the must gather path.
// It will return the Southbound Database to node mapping.
func (s *MustGatherMCPServer) ListSouthboundDatabases(ctx context.Context, req *mcp.CallToolRequest, in types.ListDatabasesParams) (*mcp.CallToolResult, types.ListDatabasesResult, error) {
	if s.ovsdbTool == nil {
		return nil, types.ListDatabasesResult{}, fmt.Errorf("ovsdb-tool is not available; ensure ovsdb-tool binary is in PATH")
	}
	output, err := s.ovsdbTool.ListSouthboundDatabases(ctx, in.MustGatherPath)
	if err != nil {
		return nil, types.ListDatabasesResult{}, err
	}
	return nil, types.ListDatabasesResult{Data: output}, nil
}

// QueryDatabase queries the database for the given database name, table, where, and columns.
// It will return the result of the query.
func (s *MustGatherMCPServer) QueryDatabase(ctx context.Context, req *mcp.CallToolRequest, in types.QueryDatabaseParams) (*mcp.CallToolResult, types.QueryDatabaseResult, error) {
	if s.ovsdbTool == nil {
		return nil, types.QueryDatabaseResult{}, fmt.Errorf("ovsdb-tool is not available; ensure ovsdb-tool binary is in PATH")
	}
	output, err := s.ovsdbTool.QueryDatabase(ctx, in.MustGatherPath, in.DatabaseName, in.Table, in.Conditions, in.Columns)
	if err != nil {
		return nil, types.QueryDatabaseResult{}, err
	}
	return nil, types.QueryDatabaseResult{Data: output}, nil
}
