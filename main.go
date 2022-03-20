package main

import (
	"context"
	"dc-top/docker"
	"dc-top/docker/compose"
	"dc-top/gui"
	"dc-top/gui/view/window"
	"dc-top/logger"
	"flag"
	"fmt"
	"log"
	"os"
)

// TODO: replace as many log.Fatal with error handling

func main() {
	const workdir = "/tmp/dc-top-files"
	os.RemoveAll(workdir)
	err := os.Mkdir(workdir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(workdir)
	logger.Init()

	dc_file_path := flag.String("dc-file-path", "", "path of docker-compose.yaml file")
	flag.Parse()

	if *dc_file_path != "" {
		if err := compose.Init(context.Background(), workdir, *dc_file_path); err != nil {
			fmt.Println(err)
			return
		}
	}

	docker.Init()
	window.InitScreen()
	gui.Draw()
	window.CloseScreen()
	docker.Close()

}
