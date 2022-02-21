package gui_events

import "time"

type ChangeToLogsShellEvent struct {
	t           time.Time
	ContainerId string
}

func (e ChangeToLogsShellEvent) When() time.Time {
	return e.t
}

func NewChangeToLogsShellEvent(container_id string) ChangeToLogsShellEvent {
	return ChangeToLogsShellEvent{
		t:           time.Now(),
		ContainerId: container_id,
	}
}
