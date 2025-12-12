package types

// ListDatabasesParams is a type that contains the must gather path.
type ListDatabasesParams struct {
	MustGatherParams
}

// ListDatabasesResult is a type that contains the list of databases.
type ListDatabasesResult struct {
	Data string `json:"data"`
}

// QueryDatabaseParams is a type that contains the must gather path, database name, table, where, and columns.
type QueryDatabaseParams struct {
	MustGatherParams
	DatabaseName string   `json:"database_name"`
	Table        string   `json:"table"`
	Conditions   []string `json:"conditions,omitempty"`
	Columns      []string `json:"columns,omitempty"`
}

// QueryDatabaseResult is a type that contains the result of the query.
type QueryDatabaseResult struct {
	Data string `json:"data"`
}
