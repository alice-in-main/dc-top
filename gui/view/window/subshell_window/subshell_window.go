package subshell_window

import (
	"context"
	"dc-top/docker"
	"dc-top/gui/view/window"
	"fmt"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/gdamore/tcell/v2"
)

const output_len = 2000

type SubshellWindow struct {
	window_ctx    context.Context
	window_cancel context.CancelFunc

	is_enabled           bool
	dimensions_generator func() window.Dimensions
	resize_ch            chan interface{}
	draw_request_ch      chan interface{}
	enable_toggle        chan bool

	id              string
	highjacked_conn *types.HijackedResponse

	output        [output_len]string
	output_offset int64
}

func NewSubshellWindow(id string) SubshellWindow {
	return SubshellWindow{
		is_enabled: true,
		dimensions_generator: func() window.Dimensions {
			w, h := window.GetScreen().Size()
			return window.NewDimensions(0, 0, w-1, h-1, false)
		},
		resize_ch:       make(chan interface{}),
		draw_request_ch: make(chan interface{}),
		enable_toggle:   make(chan bool),
		id:              id,
	}
}

func (w *SubshellWindow) Open(view_ctx context.Context) {
	var err error
	w.window_ctx, w.window_cancel = context.WithCancel(view_ctx)
	w.highjacked_conn, err = docker.OpenShell(w.id, w.window_ctx, "sh")
	if err != nil {
		log.Println("Failed to open shell")
		return
	}
	go w.main()
}

func (w *SubshellWindow) Resize() {
	w.resize_ch <- nil
}

func (w *SubshellWindow) KeyPress(ev tcell.EventKey) {
	w.handleKeyEvent(&ev)
}

func (w *SubshellWindow) MousePress(_ tcell.EventMouse) {}

func (w *SubshellWindow) HandleEvent(interface{}, window.WindowType) (interface{}, error) {
	panic(1)
}

func (w *SubshellWindow) Disable() {
	log.Printf("Disable SubshellWindow...")
	w.enable_toggle <- false
}

func (w *SubshellWindow) Enable() {
	log.Printf("Enable SubshellWindow...")
	w.enable_toggle <- true
}

func (w *SubshellWindow) Close() {
	w.window_cancel()
	w.highjacked_conn.Close()
}

func (w *SubshellWindow) main() {
	window.GetScreen().Clear()
	window.GetScreen().Show()
	go w.shellReader()
	go w.shellDrawer()
}

func (w *SubshellWindow) shellReader() {
	defer window.GetScreen().PostEvent(window.NewReturnUpperViewEvent())

	for {
		select {
		case <-w.window_ctx.Done():
			return
		default:
		}
		var buff [1024]byte
		n, err := w.highjacked_conn.Reader.Read(buff[:])
		if err != nil {
			log.Printf("Stopped drawing. got error '%s'", err)
			return
		}
		new_output := string(buff[:n])
		w.output[w.output_offset] += string(new_output)

		lines := strings.Split(w.output[w.output_offset], "\n")
		for _, line := range lines {
			index := w.output_offset % output_len
			w.output[index] = line
			w.output_offset++
		}

		w.draw_request_ch <- nil
	}
}

func (w *SubshellWindow) shellDrawer() {
	for {
		select {
		case <-w.draw_request_ch:
			w.draw()
		case <-w.resize_ch:
			w.draw()
		case w.is_enabled = <-w.enable_toggle:
			w.draw()
		case <-w.window_ctx.Done():
			return
		}
	}
}

func (w *SubshellWindow) draw() {
	if w.is_enabled {
		dimensions := w.dimensions_generator()
		start_y := w.output_offset + 1 - window.Height(&dimensions)
		if start_y < 0 {
			start_y = 0
		}

		// for y := start_y; y < w.output_offset+1; y++ {
		// 	fmt.Println(w.output[y])
		// }
		text := strings.Join(w.output[start_y:w.output_offset+1], "\n")
		fmt.Print(text)
		log.Print(text)

		// 	dimensions := w.dimensions_generator()
		// 	drawer := func(x, y int) (rune, tcell.Style) {
		// 		row := w.output[y]
		// 		if x < len(row) && y < len(w.output) {
		// 			return rune(w.output[y][x]), tcell.StyleDefault
		// 		} else {
		// 			return '\x00', tcell.StyleDefault
		// 		}
		// 	}
		// 	window.DrawContents(&dimensions, drawer)
		// 	window.GetScreen().Show()
	}
}
