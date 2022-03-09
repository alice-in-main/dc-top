package elements

import (
	"github.com/gdamore/tcell/v2"
)

type TextBox struct {
	text          string
	cursor_pos    int
	prefix        StringStyler
	prefix_len    int
	default_style tcell.Style
	cursor_style  tcell.Style
}

func NewTextBox(prefix StringStyler, prefix_len int, default_style tcell.Style, cursor_style tcell.Style) TextBox {
	return TextBox{
		text:          "",
		cursor_pos:    0,
		prefix:        prefix,
		prefix_len:    prefix_len,
		default_style: default_style,
		cursor_style:  cursor_style,
	}
}

func (box *TextBox) WriteRune(r rune) {
	box.text = box.text[:box.cursor_pos] + string(r) + box.text[box.cursor_pos:]
	box.cursor_pos++
}

func (box *TextBox) Reset() {
	box.text = ""
	box.cursor_pos = 0
}

func (box *TextBox) MoveLeft() {
	if box.cursor_pos > 0 {
		box.cursor_pos--
	}
}

func (box *TextBox) MoveRight() {
	if box.cursor_pos < len(box.text) {
		box.cursor_pos++
	}
}

func (box *TextBox) Delete() {
	if box.cursor_pos < len(box.text) {
		box.text = box.text[:box.cursor_pos] + box.text[box.cursor_pos+1:]
	}
}

func (box *TextBox) Backspace() {
	if box.cursor_pos > 0 {
		box.text = box.text[:box.cursor_pos-1] + box.text[box.cursor_pos:]
		box.cursor_pos--
	}
}

func (box *TextBox) Value() string {
	return box.text
}

func (box *TextBox) HandleKey(ev *tcell.EventKey) {
	key := ev.Key()
	switch key {
	case tcell.KeyRune:
		box.WriteRune(ev.Rune())
	case tcell.KeyLeft:
		box.MoveLeft()
	case tcell.KeyRight:
		box.MoveRight()
	case tcell.KeyBackspace:
		box.Backspace()
	case tcell.KeyBackspace2:
		box.Backspace()
	case tcell.KeyDelete:
		box.Delete()
	case tcell.KeyCtrlA:
		box.Reset()
	}
}

func (box *TextBox) Style() StringStyler {
	text_styler := func(i int) (rune, tcell.Style) {
		var r rune = ' '
		var s tcell.Style = box.default_style

		if i < len(box.text) {
			r = rune(box.text[i])
		}

		if i == box.cursor_pos {
			s = box.cursor_style
		}
		return r, s
	}

	return box.prefix.Concat(box.prefix_len, text_styler)
}
