package container_logs_window

type LogContainer interface {
	Put(_log *containerLog, index int)
	Search(substr string) []int
	Get(index int) *containerLog
}

func emptyLog() containerLog {
	return newLog("", true)
}
