package gui

import (
	"context"
	"dc-top/docker"
	"dc-top/docker/compose"
	"dc-top/gui/view/window"
	"dc-top/logger"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"go.uber.org/goleak"
)

func init() {
	logger.Init()
	docker.Init()
	docker_info, err := docker.GetDockerInfo(context.Background())
	if err != nil {
		panic(err)
	}
	if docker_info.Info.Containers < 2 {
		panic("Less than 2 containers exist, can't run tests")
	}
	docker.Close()
}

func TestMain(m *testing.M) {
	defer goleak.VerifyTestMain(m)
	stop_signal := beforeEach()
	m.Run()
	afterEach(stop_signal)
}

/*
	- init docker client
	- init screen
	- start drawing
	- wait for the drawer to be ready and return
	- send a nil in the `stop_signal` channel when stopped drawing
*/
func beforeEach() (stop_signal chan interface{}) {
	docker.Init()
	window.InitScreen()
	stop_signal = make(chan interface{})
	go func() {
		Draw()
		time.Sleep(time.Second) // make sure all windows finished drawing after draw finished
		stop_signal <- nil
	}()
	readiness := NewGuiReadinessCheck(make(chan interface{}))
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
	window.CloseScreen()
	docker.Close()
}

func TestLeaksNoActions(t *testing.T) {
	tryPause()
}

func TestLeaksHelpWindow(t *testing.T) {
	toggleHelp()
	tryPause()
}

func TestLeaksInspect(t *testing.T) {
	sendDown()
	tryPause()
	toggleInspect()
	tryPause()
	toggleInspect()
	tryPause()
}

func TestLeaksLogs(t *testing.T) {
	sendUp()
	tryPause()
	toggleLogs()
	tryPause()
	toggleLogs()
	tryPause()
}

func TestLeaksLogHelpWindow(t *testing.T) {
	sendUp()
	toggleLogs()
	tryPause()
	toggleHelp()
	tryPause()
	toggleHelp()
	tryPause()
	toggleLogs()
	tryPause()
}

func TestLeaksEmptySearch(t *testing.T) {
	sendUp()
	startSearch()
	tryPause()
	typeString("!hello!") // the '!' makes sure all containers will be filtered out
	tryPause()
	enter()
	tryPause()
	sendUp()
	tryPause()
	clearSearch()
	tryPause()
}

func TestLeaksLogSearch(t *testing.T) {
	sendUp()
	toggleLogs()
	tryPause()
	startSearch()
	typeString("test")
	enter()
	tryPause()
	nextSearchResult()
	tryPause()
	clearSearch()
	tryPause()
	toggleLogs()
	tryPause()
}

func TestLeaksContainerSubshell(t *testing.T) {
	sendUp()
	enterSubshell()
	tryPause()
	typeString("ls")
	enter()
	tryPause()
	sendEof()
	tryPause()
}

func TestLeaksEdittorNoSave(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	compose.Init(ctx, "../example_dc.yaml")

	tryPause()
	enterEdittor()
	enter()
	typeString("Hello, world!")
	enter()
	tryPause()
	quitEdittorWithoutSaving()
	tryPause()

	cancel()
}

func TestLeaksEdittorSave(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	compose.Init(ctx, "../example_dc.yaml")

	tryPause()
	enterEdittor()
	tryPause()
	typeString("Hello, world!")
	enter()
	tryPause()
	sendUp()
	deleteLineInEdittor()
	saveEdittor()
	tryPause()

	cancel()
}
