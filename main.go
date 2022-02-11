package main

import (
	"dc-top/docker"
	"dc-top/gui"
	"dc-top/logger"
)

func main() {
	logger.Init()
	docker.Init()

	gui.Draw()
}
