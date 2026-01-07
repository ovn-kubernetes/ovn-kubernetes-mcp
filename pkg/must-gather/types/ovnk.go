package types

// InfoType is a type that contains the info type.
type InfoType string

const (
	InfoTypeExtraInfo   InfoType = "extrainfo"
	InfoTypeHostNetInfo InfoType = "hostnetinfo"
	InfoTypeSubnets     InfoType = "subnets"
)

// GetOvnKInfoParams is a type that contains the must gather path and info type.
type GetOvnKInfoParams struct {
	MustGatherPath string   `json:"must_gather_path"`
	InfoType       InfoType `json:"info_type"`
}

// GetOvnKInfoResult is a type that contains the ovnk info data.
type GetOvnKInfoResult struct {
	Data string `json:"data"`
}
