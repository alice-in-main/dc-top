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
)

// TODO: replace as many log.Fatal with error handling
// TODO: add scrolling to error window

func main() {
	var err error

	logger.Init()
	defer logger.Cleanup()

	dc_file_path := flag.String("f", "", "path of docker-compose.yaml file")
	flag.Parse()

	if *dc_file_path != "" {
		if err = compose.Init(context.Background(), *dc_file_path); err != nil {
			fmt.Println(err)
			return
		}
	}
	defer compose.Cleanup()

	docker.Init()
	window.InitScreen()
	err = gui.Draw()
	window.CloseScreen()
	docker.Close()
	if err != nil {
		fmt.Println(err)
	}
}
