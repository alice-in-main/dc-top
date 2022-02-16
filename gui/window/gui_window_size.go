package window

import "github.com/gdamore/tcell/v2"

func ContainerWindowSize(s tcell.Screen) (x1, y1, x2, y2 int) {
	width, height := s.Size()
	return 1, 1, width - 2, int(0.7*float64(height) - 1)
}

func DockerInfoWindowSize(s tcell.Screen) (x1, y1, x2, y2 int) {
	width, height := s.Size()
	return 1, int(0.7*float64(height) + 1), width - 2, height - 1
}

func ContainerLogWindowSize(s tcell.Screen) (x1, y1, x2, y2 int) {
	width, height := s.Size()
	return 0, 0, width, height
}
