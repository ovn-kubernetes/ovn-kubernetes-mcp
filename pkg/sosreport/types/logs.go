package types

// SearchPodLogsParams are the parameters for sos-search-pod-logs
type SearchPodLogsParams struct {
	SosreportPath string `json:"sosreport_path"`
	Pattern       string `json:"pattern"`
	PodFilter     string `json:"pod_filter,omitempty"`
	MaxResults    int    `json:"max_results,omitempty"`
}

// SearchPodLogsResult returns matching pod log lines
type SearchPodLogsResult struct {
	Output string `json:"output"`
}
