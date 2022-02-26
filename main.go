package main

import (
	"dc-top/docker"
	"dc-top/docker/compose"
	"dc-top/gui"
	"dc-top/logger"
	"flag"
	"log"
	"os"
)

// type logsWriterr struct {
// 	logs              [docker.NumSavedLogs][]byte
// 	inner_write_index int
// }

// func (w *logsWriterr) Write(log_line []byte) (int, error) {
// 	var nl_index int
// 	for offset := 0; nl_index != -1 && offset < len(log_line); offset += (nl_index + 1) {
// 		nl_index = utils.FindByte('\n', []byte(log_line[offset:]))
// 		if nl_index != -1 {
// 			w.logs[w.inner_write_index] = log_line[offset : offset+nl_index]
// 		} else {
// 			w.logs[w.inner_write_index] = log_line[offset:]
// 		}
// 		fmt.Println(string(w.logs[w.inner_write_index]))
// 		w.inner_write_index = (w.inner_write_index + 1) % docker.NumSavedLogs
// 	}
// 	return len(log_line), nil
// }

func main() {
	const workdir = "/tmp/dc-top-files"
	err := os.Mkdir(workdir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(workdir)
	logger.Init(workdir)

	dc_enabled := flag.Bool("dc-mode", true, "docker-compose mode")
	dc_file_path := flag.String("dc-file-path", "./docker-compose.yaml", "path of docker-compose.yaml file")
	flag.Parse()

	if *dc_enabled {
		compose.Init(workdir, *dc_file_path)
	}
	docker.Init()
	gui.Draw()

	// tty, err := tty.Open()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer tty.Close()

	// for {
	// 	r, err := tty.ReadRune()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println(([]byte(string(r))))
	// }

	// writer := logsWriterr{
	// 	inner_write_index: 0,
	// }
	// docker.StreamContainerLogs("d9bf6a5c5457", &writer, context.Background())

	// x := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	// fmt.Println(utils.CutString(x, 2))
	// fmt.Println(utils.CutString(x, 3))
	// fmt.Println(utils.CutString(x, 5))
	// fmt.Println(utils.CutString(x, 0))

	// exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()

	// var b []byte = make([]byte, 1)
	// for {
	// 	os.Stdin.Read(b)
	// 	fmt.Println("I got the byte", b, "("+string(b)+")")
	// }

}
