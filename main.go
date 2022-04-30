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
	"os"
)

func main() {
	var err error

	logger.Init()
	defer logger.Cleanup()

	dc_file_path := flag.String("f", "", "path of docker-compose.yaml file")
	flag.Parse()

	if *dc_file_path != "" {
		if err = compose.Init(context.Background(), *dc_file_path); err != nil {
			fmt.Println(err)
			if _, err := os.Stat(*dc_file_path); err == nil {
				out, _ := compose.Config(context.Background())
				fmt.Print(string(out))
			}
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
