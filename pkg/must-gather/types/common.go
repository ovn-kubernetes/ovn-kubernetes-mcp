package types

// MustGatherParams is a type that contains the must gather path.
type MustGatherParams struct {
	// The path to the must gather directory. It must be an absolute path
	// and contain the must-gather.log file.
	MustGatherPath string `json:"must_gather_path"`
}
