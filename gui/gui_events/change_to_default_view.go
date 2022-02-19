package gui_events

import "time"

type ChangeToDefaultViewEvent struct {
	t time.Time
}

func (e ChangeToDefaultViewEvent) When() time.Time {
	return e.t
}

func NewChangeToDefaultViewEvent() ChangeToDefaultViewEvent {
	return ChangeToDefaultViewEvent{
		t: time.Now(),
	}
}
