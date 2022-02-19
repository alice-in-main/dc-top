package gui_events

import "time"

type ChangeToLogsWindowEvent struct {
	t           time.Time
	ContainerId string
}

func (e ChangeToLogsWindowEvent) When() time.Time {
	return e.t
}

func NewChangeToLogsWindowEvent(container_id string) ChangeToLogsWindowEvent {
	return ChangeToLogsWindowEvent{
		t:           time.Now(),
		ContainerId: container_id,
	}
}
