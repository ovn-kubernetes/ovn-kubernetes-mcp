package ovsdbtool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"

	k8sTypes "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/types"
)

const (
	networkLogsDirectory     = "network_logs"
	ovnkDatabaseStore        = "ovnk_database_store"
	ovnkubeNodeLabelSelector = "app=ovnkube-node"
)

// ListNorthboundDatabases lists the Northbound Databases in the must gather path.
// It will return the Northbound Database to node mapping.
func (s *OvsdbTool) ListNorthboundDatabases(ctx context.Context, mustGatherPath string) (string, error) {
	// Get the Northbound Database to node mapping
	dbToNode, err := s.getDatabaseToNodeMapping(ctx, mustGatherPath, true)
	if err != nil {
		return "", fmt.Errorf("failed to get database to node mapping: %w", err)
	}
	// Return the Northbound Database to node mapping
	return dbToNode, nil
}

// ListSouthboundDatabases lists the Southbound Databases in the must gather path.
// It will return the Southbound Database to node mapping.
func (s *OvsdbTool) ListSouthboundDatabases(ctx context.Context, mustGatherPath string) (string, error) {
	// Get the Southbound Database to node mapping
	dbToNode, err := s.getDatabaseToNodeMapping(ctx, mustGatherPath, false)
	if err != nil {
		return "", fmt.Errorf("failed to get database to node mapping: %w", err)
	}
	// Return the Southbound Database to node mapping
	return dbToNode, nil
}

// getDatabaseToNodeMapping gets the database to node mapping for the given must gather path and isNorthbound.
// It will return the database to node mapping.
func (s *OvsdbTool) getDatabaseToNodeMapping(ctx context.Context, mustGatherPath string, isNorthbound bool) (string, error) {
	// Get the list of all the databases
	dbFiles, err := listOvnDatabases(mustGatherPath, isNorthbound)
	if err != nil {
		return "", fmt.Errorf("failed to list ovn databases: %w", err)
	}

	if len(dbFiles) == 0 {
		return "", fmt.Errorf("no ovn databases found")
	}

	// Get the list of all the pods
	data, err := s.omcClient.ListResources(ctx, mustGatherPath, "pod", "", ovnkubeNodeLabelSelector, k8sTypes.JSONOutputType)
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %w", err)
	}

	// Parse the list of pods
	pods := &corev1.PodList{}
	err = json.Unmarshal([]byte(data), pods)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal ovnkube-node pods data: %w", err)
	}

	// Get the database to node mapping. The database files are named as <pod-name>_<nbdb|sbdb>.
	// Thus, iterate over the list of pods and find the database file for the pod. There should
	// always be one pod corresponding to each database file.
	type dbInfo struct {
		Database string `json:"database"`
		Node     string `json:"node"`
	}
	dbToNodeData := []dbInfo{}
	// Iterate over the list of pods and find the database file for the pod.
	for _, pod := range pods.Items {
		// If there are no more database files to process, break the loop
		if len(dbFiles) == 0 {
			break
		}
		// Get the database file for the pod
		dbFile, ok := dbFiles[pod.Name]
		if !ok {
			continue
		}
		// Get the node name of the pod
		nodeName := pod.Spec.NodeName
		// Skip pods that haven't been scheduled to a node yet
		if nodeName == "" {
			log.Printf("WARNING: pod %s has no node assignment, skipping", pod.Name)
			continue
		}
		// Add the pod data to the database to node data
		dbToNodeData = append(dbToNodeData, dbInfo{
			Database: dbFile,
			Node:     nodeName,
		})
		// Remove entry for the database file from the map
		delete(dbFiles, pod.Name)
	}

	// If there are still database files left, log a warning.
	if len(dbFiles) > 0 {
		log.Printf("WARNING: found %d database files left after matching pods: %v", len(dbFiles), dbFiles)
	}

	// Convert the database to node data to a json string
	dbToNode, err := json.Marshal(dbToNodeData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal database to node data: %w", err)
	}

	return string(dbToNode), nil
}

// listOvnDatabases lists the ovn databases in the must gather path. It will return the
// pod name to database file name mapping.
func listOvnDatabases(mustgatherPath string, isNorthbound bool) (map[string]string, error) {
	// Extract the ovn databases from the must gather path
	dbPath, err := extractOvnDatabases(mustgatherPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract ovn databases: %w", err)
	}

	// Read the ovn databases from the database path
	dbFiles, err := os.ReadDir(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ovn databases: %w", err)
	}

	// List the ovn databases
	dbFileNames := map[string]string{}
	for _, dbFile := range dbFiles {
		if !dbFile.IsDir() {
			// Get the pod name from the database file name. The database file name is like:
			// <pod-name>_<nbdb|sbdb>.
			lastIndex := strings.LastIndex(dbFile.Name(), "_")
			if lastIndex == -1 || lastIndex == 0 {
				log.Printf("WARNING: skipping file with unexpected name format: %s", dbFile.Name())
				continue
			}
			podName := dbFile.Name()[:lastIndex]
			// If the database is a northbound database, add it to the list
			if isNorthbound {
				if strings.HasSuffix(dbFile.Name(), "_nbdb") {
					dbFileNames[podName] = dbFile.Name()
				}
			} else {
				// If the database is a southbound database, add it to the list
				if strings.HasSuffix(dbFile.Name(), "_sbdb") {
					dbFileNames[podName] = dbFile.Name()
				}
			}
		}
	}
	return dbFileNames, nil
}
