package window

import "github.com/gdamore/tcell/v2"

func ContainerWindowSize(s tcell.Screen) (x1, y1, x2, y2 int) {
	width, height := s.Size()
	return 1, 1, width - 2, int(0.7 * float64(height))
}

func ContainersBarWindowSize(s tcell.Screen) (x1, y1, x2, y2 int) {
	width, height := s.Size()
	return 1, int(0.7 * float64(height)), width - 2, int(0.7*float64(height) + 2)
}

func DockerInfoWindowSize(s tcell.Screen) (x1, y1, x2, y2 int) {
	width, height := s.Size()
	return 1, int(0.7*float64(height) + 2), (width - 2) / 2, height - 1
}
