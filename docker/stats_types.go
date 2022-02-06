package docker

type ContainerMainStats struct {
	Name   string      `json:"name"`
	Cpu    CpuStats    `json:"cpu_stats"`
	PreCpu CpuStats    `json:"precpu_stats"`
	Memory MemoryStats `json:"memory_stats"`
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
