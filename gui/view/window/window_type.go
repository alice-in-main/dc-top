package window

type WindowType uint8

const (
	ContainersHolder WindowType = iota
	DockerInfo
	Bar
	GeneralInfo
	ContainerLogs
	Help
	Edittor
	Subshell
	Other
)
