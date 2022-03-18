package window

func GeneralInfoWindowSize() (x1, y1, x2, y2 int) {
	width, _ := GetScreen().Size()
	return 2, 0, width - 2, 1
}

func ContainerWindowSize() (x1, y1, x2, y2 int) {
	width, height := GetScreen().Size()
	return 1, 4, width - 2, int(0.7 * float64(height))
}

func ContainersBarWindowSize() (x1, y1, x2, y2 int) {
	width, height := GetScreen().Size()
	return 2, int(0.7*float64(height)) + 1, width - 2, int(0.7*float64(height) + 1)
}

func DockerInfoWindowSize() (x1, y1, x2, y2 int) {
	width, height := GetScreen().Size()
	return 1, int(0.7*float64(height) + 2), (width-2)/2 - 1, height - 1
}

func MainHelpWindowSize() (x1, y1, x2, y2 int) {
	width, height := GetScreen().Size()
	return (width-2)/2 + 1, int(0.7*float64(height) + 2), (width - 2), height - 1
}

func LogsWindowSize() (x1, y1, x2, y2 int) {
	width, height := GetScreen().Size()
	return 0, 0, width, height - 2
}

func LogsBarWindowSize() (x1, y1, x2, y2 int) {
	width, height := GetScreen().Size()
	return 0, height - 1, width, height - 1
}
