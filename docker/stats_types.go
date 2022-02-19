package docker

import "time"

// STATS
type ContainerMainStats struct {
	Name       string                  `json:"name"`
	Cpu        CpuStats                `json:"cpu_stats"`
	PreCpu     CpuStats                `json:"precpu_stats"`
	Memory     MemoryStats             `json:"memory_stats"`
	Network    map[string]NetworkUsage `json:"networks"`
	PreNetwork map[string]NetworkUsage
}

type CpuStats struct {
	ContainerUsage CpuUsage `json:"cpu_usage"`
	SystemUsage    int64    `json:"system_cpu_usage"`
}

type MemoryStats struct {
	Usage int64 `json:"usage"`
	Limit int64 `json:"limit"`
}

type CpuUsage struct {
	TotalUsage int64 `json:"total_usage"`
}

type NetworkUsage struct {
	ReceivedBytes      uint64    `json:"rx_bytes"`
	ReceivedPackets    uint64    `json:"rx_packets"`
	ReceivedErrors     uint64    `json:"rx_errors"`
	ReceivedDropped    uint64    `json:"rx_dropped"`
	TransmittedBytes   uint64    `json:"tx_bytes"`
	TransmittedPackets uint64    `json:"tx_packets"`
	TransmittedErrors  uint64    `json:"tx_errors"`
	TransmittedDropped uint64    `json:"tx_dropped"`
	LastUpdateTime     time.Time `json:"-"`
}

// STATS END
