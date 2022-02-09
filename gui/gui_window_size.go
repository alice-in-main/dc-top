package gui

import "github.com/gdamore/tcell/v2"

func containerWindowSize(s tcell.Screen) (x1, y1, x2, y2 int) {
	width, height := s.Size()
	return 1, 1, width - 2, int(0.7*float64(height) - 1)
}

func dockerInfoWindowSize(s tcell.Screen) (x1, y1, x2, y2 int) {
	width, height := s.Size()
	return 1, int(0.7*float64(height) + 1), width - 2, height - 1
}
