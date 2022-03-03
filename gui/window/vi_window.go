package window

import (
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/gdamore/tcell/v2"
)

type ViWindow struct {
	file_path string
	context   context.Context
	sender    WindowType
}

func NewViWindow(file_path string, sender WindowType, context context.Context) ViWindow {
	return ViWindow{
		file_path: file_path,
		context:   context,
		sender:    sender,
	}
}

func (w *ViWindow) Open(s tcell.Screen) {
	go w.main(s)
}

func (w *ViWindow) Resize() {}

func (w *ViWindow) KeyPress(_ tcell.EventKey) {}

func (w *ViWindow) MousePress(_ tcell.EventMouse) {}

func (w *ViWindow) HandleEvent(interface{}, WindowType) (interface{}, error) {
	log.Println("Vi window got event")
	panic(1)
}

func (w *ViWindow) Close() {}

func (w *ViWindow) main(s tcell.Screen) {
	orig_last_edit_date, err := lastUpdateTime(w.file_path)
	exitIfErr(s, err)
	s.Suspend()
	possible_editors := []string{"vim", "vi"}
	for _, editor := range possible_editors {
		cmd := exec.CommandContext(w.context, editor, w.file_path)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err == nil {
			s.Resume()
			s.PostEvent(NewChangeToDefaultViewEvent())
			log.Println("resuming from vi")
			final_last_edit_date, err := lastUpdateTime(w.file_path)
			exitIfErr(s, err)
			s.PostEvent(NewMessageEvent(w.sender, Vi, ViWindowResult{
				FilePath:   w.file_path,
				WasEditted: final_last_edit_date.After(orig_last_edit_date),
			}))
			return
		}
		log.Printf("%s failed with '%s'", editor, err)
	}
	exitIfErr(s, errors.New("vi window failed"))
}

type ViWindowResult struct {
	FilePath   string
	WasEditted bool
}

func lastUpdateTime(path string) (time.Time, error) {
	file, err := os.Stat(path)
	if err != nil {
		return time.Now(), err
	}
	return file.ModTime(), nil
}
