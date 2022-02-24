package window

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

// ---------

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

// ---------

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

// ---------

type MessageEvent struct {
	t        time.Time
	Receiver WindowType
	Message  interface{}
}

func (e MessageEvent) When() time.Time {
	return e.t
}

func NewMessageEvent(receiver_window WindowType, message interface{}) MessageEvent {
	return MessageEvent{
		t:        time.Now(),
		Receiver: receiver_window,
		Message:  message,
	}
}
