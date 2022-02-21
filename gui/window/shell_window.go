package window

import (
	"context"
	"dc-top/docker"
	"dc-top/gui/gui_events"

	"github.com/gdamore/tcell/v2"
)

type ShellWindow struct {
	id      string
	context context.Context
}

func NewShellWindow(id string, context context.Context) ShellWindow {
	return ShellWindow{
		id:      id,
		context: context,
	}
}

func (w *ShellWindow) Open(s tcell.Screen) {
	go w.main(s)
}

func (w *ShellWindow) Resize() {}

func (w *ShellWindow) KeyPress(_ tcell.EventKey) {}

func (w *ShellWindow) MousePress(_ tcell.EventMouse) {}

func (w *ShellWindow) Close() {}

func (w *ShellWindow) main(s tcell.Screen) {
	defer s.PostEvent(gui_events.NewChangeToDefaultViewEvent())
	s.Suspend()
	defer s.Resume()
	possible_shells := []string{"bash", "sh"}
	for _, sh := range possible_shells {
		if docker.OpenShell(w.id, w.context, sh) == nil {
			return
		}
	}
}
