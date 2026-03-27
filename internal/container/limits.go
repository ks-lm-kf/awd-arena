package container

// ResourceLimit defines resource constraints for a container.
type ResourceLimit struct {
	CPUCores   float64 `json:"cpu_cores"`
	MemoryMB   int64   `json:"memory_mb"`
	DiskGB     int64   `json:"disk_gb"`
	NetworkBPS int64   `json:"network_bps"`
	PidsLimit  int     `json:"pids_limit"`
}

// DefaultResourceLimit returns default limits.
func DefaultResourceLimit() ResourceLimit {
	return ResourceLimit{
		CPUCores:   0.5,
		MemoryMB:   256,
		DiskGB:     1,
		NetworkBPS: 10 * 1024 * 1024,
		PidsLimit:  64,
	}
}
