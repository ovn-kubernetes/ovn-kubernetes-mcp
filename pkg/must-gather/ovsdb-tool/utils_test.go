package ovsdbtool

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestBuildQueryString(t *testing.T) {
	tests := []struct {
		testName        string
		schemaName      string
		table           string
		conditions      []string
		columns         []string
		wantQueryString string
		wantError       bool
	}{
		{
			testName:        "valid query string with one condition",
			schemaName:      "OVN_Southbound",
			table:           "Chassis",
			conditions:      []string{"[\"hostname\",\"==\",\"node-1\"]"},
			columns:         []string{"_uuid"},
			wantQueryString: "[\"OVN_Southbound\", {\"op\": \"select\", \"table\": \"Chassis\", \"where\": [[\"hostname\",\"==\",\"node-1\"]], \"columns\": [\"_uuid\"]}]",
			wantError:       false,
		},
		{
			testName:        "valid query string with one condition and no columns",
			schemaName:      "OVN_Southbound",
			table:           "Chassis",
			conditions:      []string{"[\"hostname\",\"==\",\"node-1\"]"},
			columns:         []string{},
			wantQueryString: "[\"OVN_Southbound\", {\"op\": \"select\", \"table\": \"Chassis\", \"where\": [[\"hostname\",\"==\",\"node-1\"]]}]",
			wantError:       false,
		},
		{
			testName:        "valid query string with multiple conditions and multiple columns",
			schemaName:      "OVN_Southbound",
			table:           "Chassis",
			conditions:      []string{"[\"hostname\",\"!=\",\"node-1\"]", "[\"hostname\",\"!=\",\"node-2\"]"},
			columns:         []string{"_uuid", "_version"},
			wantQueryString: "[\"OVN_Southbound\", {\"op\": \"select\", \"table\": \"Chassis\", \"where\": [[\"hostname\",\"!=\",\"node-1\"],[\"hostname\",\"!=\",\"node-2\"]], \"columns\": [\"_uuid\",\"_version\"]}]",
			wantError:       false,
		},
		{
			testName:        "valid query string with multiple conditions and no columns",
			schemaName:      "OVN_Southbound",
			table:           "Chassis",
			conditions:      []string{"[\"hostname\",\"!=\",\"node-1\"]", "[\"hostname\",\"!=\",\"node-2\"]"},
			columns:         []string{},
			wantQueryString: "[\"OVN_Southbound\", {\"op\": \"select\", \"table\": \"Chassis\", \"where\": [[\"hostname\",\"!=\",\"node-1\"],[\"hostname\",\"!=\",\"node-2\"]]}]",
			wantError:       false,
		},
		{
			testName:        "query with no conditions and multiple columns",
			schemaName:      "OVN_Southbound",
			table:           "Chassis",
			conditions:      []string{},
			columns:         []string{"_uuid", "_version"},
			wantQueryString: "[\"OVN_Southbound\", {\"op\": \"select\", \"table\": \"Chassis\", \"where\": [], \"columns\": [\"_uuid\",\"_version\"]}]",
			wantError:       false,
		},
		{
			testName:        "query with no conditions and no columns",
			schemaName:      "OVN_Southbound",
			table:           "Chassis",
			conditions:      []string{},
			columns:         []string{},
			wantQueryString: "[\"OVN_Southbound\", {\"op\": \"select\", \"table\": \"Chassis\", \"where\": []}]",
			wantError:       false,
		},
		{
			testName:        "invalid schema name",
			schemaName:      "invalid",
			table:           "Chassis",
			conditions:      []string{},
			columns:         []string{},
			wantQueryString: "",
			wantError:       true,
		},
		{
			testName:        "invalid table",
			schemaName:      "OVN_Southbound",
			table:           "",
			conditions:      []string{},
			columns:         []string{},
			wantQueryString: "",
			wantError:       true,
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			queryString, err := buildQueryString(test.schemaName, test.table, test.conditions, test.columns)
			if err != nil && !test.wantError {
				t.Fatalf("buildQueryString() error = %v, want nil", err)
			}
			if err == nil && test.wantError {
				t.Fatalf("buildQueryString() expected error but got nil")
			}
			if err == nil {
				var queryJson, wantQueryJson []any
				err = json.Unmarshal([]byte(queryString), &queryJson)
				if err != nil {
					t.Fatalf("failed to unmarshal query to map: %v", err)
				}
				err = json.Unmarshal([]byte(test.wantQueryString), &wantQueryJson)
				if err != nil {
					t.Fatalf("failed to unmarshal want query to map: %v", err)
				}
				if !reflect.DeepEqual(queryJson, wantQueryJson) {
					t.Fatalf("queryJson and wantQueryJson are not equal: queryJson: %v, wantQueryJson: %v", queryJson, wantQueryJson)
				}
			}
		})
	}
}
