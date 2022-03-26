package window

import (
	"log"
	"time"
)

// type ChangeToDefaultViewEvent struct {
// 	t time.Time
// }

// func (e ChangeToDefaultViewEvent) When() time.Time {
// 	return e.t
// }

// func NewChangeToDefaultViewEvent() ChangeToDefaultViewEvent {
// 	return ChangeToDefaultViewEvent{
// 		t: time.Now(),
// 	}
// }

// ---------

type ReturnUpperViewEvent struct {
	t time.Time
}

func (e ReturnUpperViewEvent) When() time.Time {
	return e.t
}

func NewReturnUpperViewEvent() ReturnUpperViewEvent {
	return ReturnUpperViewEvent{
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

type ChangeToMainHelpEvent struct {
	t time.Time
}

func (e ChangeToMainHelpEvent) When() time.Time {
	return e.t
}

func NewChangeToMainHelpEvent() ChangeToMainHelpEvent {
	return ChangeToMainHelpEvent{
		t: time.Now(),
	}
}

// ---------

type ChangeToLogsHelpEvent struct {
	t time.Time
}

func (e ChangeToLogsHelpEvent) When() time.Time {
	return e.t
}

func NewChangeToLogsHelpEvent() ChangeToLogsHelpEvent {
	return ChangeToLogsHelpEvent{
		t: time.Now(),
	}
}

// ---------

type ChangeToEdittorHelpEvent struct {
	t time.Time
}

func (e ChangeToEdittorHelpEvent) When() time.Time {
	return e.t
}

func NewChangeToEdittorHelpEvent() ChangeToEdittorHelpEvent {
	return ChangeToEdittorHelpEvent{
		t: time.Now(),
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

func ExitIfErr(err error) {
	if err != nil && err.Error() != "context canceled" {
		log.Printf("a fatal error occured: %s\n", err)
		GetScreen().PostEvent(NewFatalErrorEvent(err))
	}
}
