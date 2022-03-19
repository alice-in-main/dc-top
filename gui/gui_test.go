package gui

import (
	"dc-top/docker"
	"dc-top/gui/view/window"
	"testing"

	"github.com/gdamore/tcell/v2"
	"go.uber.org/goleak"
)

/*
	- init docker client
	- init screen
	- start drawing
	- wait for the drawer to be ready and return
	- send a nil in the `stop_signal` channel when stopped drawing
*/
func beforeEach() (stop_signal chan interface{}) {
	docker.Init()
	InitScreen()
	stop_signal = make(chan interface{})
	go func() {
		Draw()
		stop_signal <- nil
	}()
	readiness := NewReadinessCheck(make(chan interface{}))
	window.GetScreen().PostEvent(readiness)
	<-readiness.Ack
	return stop_signal
}

/*
	- send Ctrl+C to stop drawing
	- wait for `stop_signal`
	- close screen
	- close docker client
*/
func afterEach(stop_signal chan interface{}) {
	stopKey := tcell.NewEventKey(tcell.KeyCtrlC, '\x00', 0)
	window.GetScreen().PostEvent(stopKey)
	<-stop_signal
	CloseScreen()
	docker.Close()
}

func TestSanity(t *testing.T) {
	stop_signal := beforeEach()
	afterEach(stop_signal)
}

func TestLeaksNoActions(t *testing.T) {
	defer goleak.VerifyNone(t)
	stop_signal := beforeEach()
	afterEach(stop_signal)
}
