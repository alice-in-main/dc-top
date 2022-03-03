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

type ChangeToViWindowEvent struct {
	t        time.Time
	FilePath string
	Sender   WindowType
}

func (e ChangeToViWindowEvent) When() time.Time {
	return e.t
}

func NewChangeToViWindowEvent(file_path string, sender_window WindowType) ChangeToViWindowEvent {
	return ChangeToViWindowEvent{
		t:        time.Now(),
		FilePath: file_path,
		Sender:   sender_window,
	}
}

// ---------

type MessageEvent struct {
	t        time.Time
	Receiver WindowType
	Sender   WindowType
	Message  interface{}
}

func (e MessageEvent) When() time.Time {
	return e.t
}

func NewMessageEvent(receiver WindowType, sender WindowType, message interface{}) MessageEvent {
	return MessageEvent{
		t:        time.Now(),
		Receiver: receiver,
		Sender:   sender,
		Message:  message,
	}
}

// ---------

type FatalErrorEvent struct {
	t   time.Time
	Err error
}

func (e FatalErrorEvent) When() time.Time {
	return e.t
}

func NewFatalErrorEvent(err error) FatalErrorEvent {
	return FatalErrorEvent{
		t:   time.Now(),
		Err: err,
	}
}
