package main

import (
	"dc-top/docker"
	"dc-top/gui"
	"dc-top/logger"
	"dc-top/utils"
	"fmt"
)

type logsWriterr struct {
	logs              [docker.NumSavedLogs][]byte
	inner_write_index int
}

func (w *logsWriterr) Write(log_line []byte) (int, error) {
	var nl_index int
	for offset := 0; nl_index != -1 && offset < len(log_line); offset += (nl_index + 1) {
		nl_index = utils.FindByte('\n', []byte(log_line[offset:]))
		if nl_index != -1 {
			w.logs[w.inner_write_index] = log_line[offset : offset+nl_index]
		} else {
			w.logs[w.inner_write_index] = log_line[offset:]
		}
		fmt.Println(string(w.logs[w.inner_write_index]))
		w.inner_write_index = (w.inner_write_index + 1) % docker.NumSavedLogs
	}
	return len(log_line), nil
}

func main() {
	logger.Init()
	docker.Init()

	gui.Draw()

	// writer := logsWriterr{
	// 	inner_write_index: 0,
	// }
	// docker.StreamContainerLogs("d9bf6a5c5457", &writer, context.Background())
}
