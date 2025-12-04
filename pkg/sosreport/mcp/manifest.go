package sosreport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/sosreport/types"
)

// loadManifest loads and parses the manifest.json file
func loadManifest(sosreportPath string) (*types.Manifest, error) {
	manifestPath := filepath.Join(sosreportPath, "sos_reports", "manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest.json: %w", err)
	}

	var manifest types.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest.json: %w", err)
	}

	return &manifest, nil
}
