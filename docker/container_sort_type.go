package docker

type SortType uint8

const (
	Name SortType = iota
	Memory
	Cpu
	Image
	State
	None
)

var typeToString = map[SortType]string{
	Name:   "Name",
	Memory: "Memory",
	Cpu:    "CPU",
	Image:  "Image",
	State:  "State",
}

func (sort_type SortType) String() string {
	return typeToString[sort_type]
}
