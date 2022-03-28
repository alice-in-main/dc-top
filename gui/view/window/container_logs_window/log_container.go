package container_logs_window

type LogContainer interface {
	Put(_log *singleLog, index int)
	Search(substr string) []int
	Get(index int) *singleLog
}

func emptyLog() singleLog {
	return newLog("", true)
}
