package types

// GetOvnKInfoParams is a type that contains the must gather path and info type.
type GetOvnKInfoParams struct {
	MustGatherPath string `json:"must_gather_path"`
	InfoType       string `json:"info_type"`
}

// GetOvnKInfoResult is a type that contains the ovnk info data.
type GetOvnKInfoResult struct {
	Data string `json:"data"`
}
