package main

import (
	"dc-top/docker"
	"dc-top/gui"
)

func main() {
	docker.Init()
	// data := docker.GetContainers()
	// fmt.Println(data[0].String())
	// fmt.Println(data[1].String())
	// fmt.Println(data[2].String())

	gui.Draw()
}
