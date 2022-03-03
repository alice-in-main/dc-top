package main

import (
	"dc-top/docker"
	"dc-top/docker/compose"
	"dc-top/gui"
	"dc-top/logger"
	"flag"
	"fmt"
	"log"
	"os"
)

// TODO: replace as many log.Fatal with error handling

func main() {

	const workdir = "/tmp/dc-top-files"
	err := os.Mkdir(workdir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(workdir)
	logger.Init()

	dc_enabled := flag.Bool("dc-mode", true, "docker-compose mode")
	dc_file_path := flag.String("dc-file-path", "./docker-compose.yaml", "path of docker-compose.yaml file")
	flag.Parse()

	if *dc_enabled {
		if err := compose.Init(workdir, *dc_file_path); err != nil {
			fmt.Println(err)
			return
		}
	}

	docker.Init()
	gui.Draw()

}
