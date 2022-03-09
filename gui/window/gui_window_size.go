package window

func GeneralInfoWindowSize() (x1, y1, x2, y2 int) {
	width, _ := GetScreen().Size()
	return 1, 0, width - 2, 4
}

func ContainerWindowSize() (x1, y1, x2, y2 int) {
	width, height := GetScreen().Size()
	return 1, 4, width - 2, int(0.7 * float64(height))
}

func ContainersBarWindowSize() (x1, y1, x2, y2 int) {
	width, height := GetScreen().Size()
	return 1, int(0.7 * float64(height)), width - 2, int(0.7*float64(height) + 2)
}

func DockerInfoWindowSize() (x1, y1, x2, y2 int) {
	width, height := GetScreen().Size()
	return 1, int(0.7*float64(height) + 2), (width - 2) / 2, height - 1
}
