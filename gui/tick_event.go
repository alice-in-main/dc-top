package gui

import "time"

type tickEvent struct{}

func (tickEvent) When() time.Time {
	return time.Now()
}
