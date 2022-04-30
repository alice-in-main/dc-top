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
	if docker_info.Info.ContainersRunning < 1 {
		panic("No running containers, can't run tests")
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
	time.Sleep(100 * time.Millisecond)
}

func TestLeaksHelpWindow(t *testing.T) {
	toggleHelp()
}

func TestLeaksInspect(t *testing.T) {
	sendDown()
	toggleInspect()
	toggleInspect()
}

func TestLeaksLogs(t *testing.T) {
	sendUp()
	toggleLogs()
	toggleLogs()
}

func TestLeaksLogHelpWindow(t *testing.T) {
	sendUp()
	sendDown()
	toggleLogs()
	toggleHelp()
	toggleHelp()
	toggleLogs()
}

func TestLeaksEmptySearch(t *testing.T) {
	sendUp()
	startSearch()
	typeString("!hello!") // the '!' makes sure all containers will be filtered out
	enter()
	sendUp()
	clearSearch()
}

func TestLeaksLogSearch(t *testing.T) {
	sendUp()
	toggleLogs()
	startSearch()
	typeString("test")
	enter()
	nextSearchResult()
	clearSearch()
	toggleLogs()
}

func TestLeaksContainerSubshell(t *testing.T) {
	sendUp()
	enterSubshell()
	typeString("ls")
	enter()
	sendEof()
}

func TestLeaksEdittorNoSave(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	if err := compose.Init(ctx, "../testutils/example_dc.yaml"); err != nil {
		t.Error(err.Error())
	}

	enterEdittor()
	typeString("Hello, world!")
	enter()
	quitEdittorWithoutSaving()

	cancel()
}

func TestLeaksEdittorSave(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	if err := compose.Init(ctx, "../testutils/example_dc.yaml"); err != nil {
		t.Error(err.Error())
	}

	enterEdittor()
	typeString("Hello, world!")
	enter()
	sendUp()
	deleteLineInEdittor()
	saveEdittor()

	cancel()
}
