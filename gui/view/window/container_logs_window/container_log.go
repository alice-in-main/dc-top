package container_logs_window

type containerLog struct {
	content   string
	is_stdout bool
}

func newLog(content string, is_stdout bool) containerLog {
	return containerLog{
		content:   content,
		is_stdout: is_stdout,
	}
}
