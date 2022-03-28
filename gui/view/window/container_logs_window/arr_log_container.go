package container_logs_window

import "strings"

type ArrLogContainer struct {
	logs []*singleLog
	size int
}

func NewArrStringSearcher(size int) ArrLogContainer {
	return ArrLogContainer{
		logs: make([]*singleLog, size),
		size: size,
	}
}

func (searcher *ArrLogContainer) Put(str *singleLog, index int) {
	searcher.logs[index] = str
}

func (searcher *ArrLogContainer) Search(substr string) []int {
	indices := make([]int, 0)
	for i, _log := range searcher.logs {
		if _log != nil && strings.Contains(_log.content, substr) {
			indices = append(indices, i)
		}
	}
	return indices
}

func (searcher *ArrLogContainer) Get(index int) *singleLog {
	if searcher.logs[index] == nil {
		empty := emptyLog()
		return &empty
	}
	return searcher.logs[index]
}
