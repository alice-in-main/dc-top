package container_logs_window

import "strings"

type ArrLogContainer struct {
	logs []*containerLog
	size int
}

func NewArrStringSearcher(size int) ArrLogContainer {
	return ArrLogContainer{
		logs: make([]*containerLog, size),
		size: size,
	}
}

func (searcher *ArrLogContainer) Put(str *containerLog, index int) {
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

func (searcher *ArrLogContainer) Get(index int) *containerLog {
	if searcher.logs[index] == nil {
		empty := emptyLog()
		return &empty
	}
	return searcher.logs[index]
}
