package window

import (
	"time"
)

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

type PauseWindowsEvent struct {
	t time.Time
}

func (e PauseWindowsEvent) When() time.Time {
	return e.t
}

func NewPauseWindowsEvent() PauseWindowsEvent {
	return PauseWindowsEvent{
		t: time.Now(),
	}
}

// ---------

type ResumeWindowsEvent struct {
	t time.Time
}

func (e ResumeWindowsEvent) When() time.Time {
	return e.t
}

func NewResumeWindowsEvent() ResumeWindowsEvent {
	return ResumeWindowsEvent{
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

type ChangeToContainerShellEvent struct {
	t           time.Time
	ContainerId string
}

func (e ChangeToContainerShellEvent) When() time.Time {
	return e.t
}

func NewChangeToContainerShellEvent(container_id string) ChangeToContainerShellEvent {
	return ChangeToContainerShellEvent{
		t:           time.Now(),
		ContainerId: container_id,
	}
}

// ---------

type ChangeToFileEdittorEvent struct {
	t        time.Time
	FilePath string
	Sender   WindowType
}

func (e ChangeToFileEdittorEvent) When() time.Time {
	return e.t
}

func NewChangeToFileEdittorEvent(file_path string, sender_window WindowType) ChangeToFileEdittorEvent {
	return ChangeToFileEdittorEvent{
		t:        time.Now(),
		FilePath: file_path,
		Sender:   sender_window,
	}
}

// ---------

type StopDrawingEvent struct {
	t time.Time
}

func (e StopDrawingEvent) When() time.Time {
	return e.t
}

func NewStopDrawingEvent(file_path string, sender_window WindowType) StopDrawingEvent {
	return StopDrawingEvent{
		t: time.Now(),
	}
}

// ---------

type StartDrawingEvent struct {
	t time.Time
}

func (e StartDrawingEvent) When() time.Time {
	return e.t
}

func NewStartDrawingEvent(file_path string, sender_window WindowType) StartDrawingEvent {
	return StartDrawingEvent{
		t: time.Now(),
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