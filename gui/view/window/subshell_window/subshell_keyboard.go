package subshell_window

import (
	"github.com/gdamore/tcell/v2"
)

func (w *SubshellWindow) handleKeyEvent(ev *tcell.EventKey) {
	key := ev.Key()
	switch key {
	case tcell.KeyEnter:
		w.highjacked_conn.Conn.Write([]byte{0xA})
	case tcell.KeyTab:
		w.highjacked_conn.Conn.Write([]byte{0x9})
	case tcell.KeyBackspace:
		w.highjacked_conn.Conn.Write([]byte{0x8}) // does nothing?
	case tcell.KeyBackspace2:
		w.highjacked_conn.Conn.Write([]byte{0x8})
	case tcell.KeyDelete:
		w.highjacked_conn.Conn.Write([]byte{27})
		w.highjacked_conn.Conn.Write([]byte{91})
		w.highjacked_conn.Conn.Write([]byte{51})
		w.highjacked_conn.Conn.Write([]byte{126})
	case tcell.KeyCtrlC:
		w.highjacked_conn.Conn.Write([]byte{0x3})
	case tcell.KeyCtrlD:
		w.highjacked_conn.Conn.Write([]byte{0x4})
	case tcell.KeyCtrlZ:
		w.highjacked_conn.Conn.Write([]byte{0x1A})
	case tcell.KeyCtrlR:
		w.highjacked_conn.Conn.Write([]byte{0x12})
	case tcell.KeyUp:
		w.highjacked_conn.Conn.Write([]byte{27})
		w.highjacked_conn.Conn.Write([]byte{79})
		w.highjacked_conn.Conn.Write([]byte{65})
	case tcell.KeyDown:
		w.highjacked_conn.Conn.Write([]byte{27})
		w.highjacked_conn.Conn.Write([]byte{79})
		w.highjacked_conn.Conn.Write([]byte{66})
	case tcell.KeyLeft:
		w.highjacked_conn.Conn.Write([]byte{27})
		w.highjacked_conn.Conn.Write([]byte{79})
		w.highjacked_conn.Conn.Write([]byte{68})
	case tcell.KeyRight:
		w.highjacked_conn.Conn.Write([]byte{27})
		w.highjacked_conn.Conn.Write([]byte{79})
		w.highjacked_conn.Conn.Write([]byte{67})
	case tcell.KeyPgUp:
		w.highjacked_conn.Conn.Write([]byte{27})
		w.highjacked_conn.Conn.Write([]byte{91})
		w.highjacked_conn.Conn.Write([]byte{53})
		w.highjacked_conn.Conn.Write([]byte{126})
	case tcell.KeyPgDn:
		w.highjacked_conn.Conn.Write([]byte{27})
		w.highjacked_conn.Conn.Write([]byte{91})
		w.highjacked_conn.Conn.Write([]byte{54})
		w.highjacked_conn.Conn.Write([]byte{126})
	case tcell.KeyHome:
		w.highjacked_conn.Conn.Write([]byte{27})
		w.highjacked_conn.Conn.Write([]byte{91})
		w.highjacked_conn.Conn.Write([]byte{49})
		w.highjacked_conn.Conn.Write([]byte{126})
	case tcell.KeyEnd:
		w.highjacked_conn.Conn.Write([]byte{27})
		w.highjacked_conn.Conn.Write([]byte{91})
		w.highjacked_conn.Conn.Write([]byte{52})
		w.highjacked_conn.Conn.Write([]byte{126})
	case tcell.KeyRune:
		a := string(ev.Rune())
		w.highjacked_conn.Conn.Write([]byte(a))
	}
}
