package main

import (
	"dc-top/docker"
	"dc-top/gui"
	"dc-top/logger"
)

func main() {
	logger.Init()
	docker.Init()

	// a := docker.GetContainers(nil)
	// for _, c := range a.GetData() {
	// 	c.Stats()
	// }
	gui.Draw()
}
