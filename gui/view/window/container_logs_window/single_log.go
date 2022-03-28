package container_logs_window

type singleLog struct {
	content   string
	is_stdout bool
}

func newLog(content string, is_stdout bool) singleLog {
	return singleLog{
		content:   content,
		is_stdout: is_stdout,
	}
}
