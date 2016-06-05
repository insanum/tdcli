
package main

import (
	"github.com/nsf/termbox-go"
	"log"
	"sync"
	"errors"
)

var ColorDefault      = int(termbox.ColorDefault)
var AttrNone          = int(termbox.ColorDefault)
var AttrBold          = int(termbox.AttrBold)
var AttrUnderline     = int(termbox.AttrUnderline)
var AttrReverse       = int(termbox.AttrReverse)

var TermKeyF1         = termbox.KeyF1
var TermKeyF2         = termbox.KeyF2
var TermKeyF3         = termbox.KeyF3
var TermKeyF4         = termbox.KeyF4
var TermKeyF5         = termbox.KeyF5
var TermKeyF6         = termbox.KeyF6
var TermKeyF7         = termbox.KeyF7
var TermKeyF8         = termbox.KeyF8
var TermKeyF9         = termbox.KeyF9
var TermKeyF10        = termbox.KeyF11
var TermKeyF11        = termbox.KeyF11
var TermKeyF12        = termbox.KeyF12

var TermKeyEscape     = termbox.KeyEsc
var TermKeyBackspace  = termbox.KeyBackspace
var TermKeyTab        = termbox.KeyTab
var TermKeyEnter      = termbox.KeyEnter
var TermKeySpace      = termbox.KeySpace
var TermKeyInsert     = termbox.KeyInsert
var TermKeyDelete     = termbox.KeyDelete
var TermKeyHome       = termbox.KeyHome
var TermKeyEnd        = termbox.KeyEnd
var TermKeyPgUp       = termbox.KeyPgup
var TermKeyPgDown     = termbox.KeyPgdn
var TermKeyArrowUp    = termbox.KeyArrowUp
var TermKeyArrowDown  = termbox.KeyArrowDown
var TermKeyArrowLeft  = termbox.KeyArrowLeft
var TermKeyArrowRight = termbox.KeyArrowRight

var TermKeyCtrlE      = termbox.KeyCtrlE
var TermKeyCtrlY      = termbox.KeyCtrlY
var TermKeyCtrlD      = termbox.KeyCtrlD
var TermKeyCtrlU      = termbox.KeyCtrlU

var mtx = &sync.Mutex{}

func TermInit() {
	err := termbox.Init()
	if err != nil {
		log.Fatal(err)
	}

	termbox.SetOutputMode(termbox.Output256)
	termbox.SetInputMode(termbox.InputEsc)
}

func TermClose() {
	termbox.Close()
}

func TermSize() (width  int,
		 height int) {
	return termbox.Size()
}

func TermDrawScreen() {
	mtx.Lock()
	termbox.Flush()
	mtx.Unlock()
}

func TermSync() {
	mtx.Lock()
	termbox.Sync()
	mtx.Unlock()
}

func TermClearScreen(full bool) {
	/*
	termbox.Clear(termbox.ColorDefault,
		      termbox.ColorDefault)
	*/

	w, h := termbox.Size()
	if !full {
		h = (h - 1)
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			termbox.SetCell(x, y, ' ',
					termbox.ColorDefault,
					termbox.ColorDefault)
		}
	}

	TermDrawScreen()
}

func TermClearLines(line int,
		    cnt  int) {
	w, h := termbox.Size()

	if (line + cnt) > h {
		h = (h - line)
	} else {
		h = (line + cnt)
	}

	for ; line < h; line++ {
		for x := 0; x < w; x++ {
			termbox.SetCell(x, line, ' ',
					termbox.ColorDefault,
					termbox.ColorDefault)
		}
	}

	TermDrawScreen()
}

func TermPrint(x   int,
	       y   int,
	       fg  interface{},
	       bg  interface{},
	       msg string) {
	for i := 0; i < len(msg); i++ {
		c := rune(msg[i])

		if c == '\n' {
			x = 0
			y++
		} else if c == '\t' {
			for j := 0; j < 4; j++ {
				termbox.SetCell(x, y, ' ',
						fg.(termbox.Attribute),
						bg.(termbox.Attribute))
				x++
			}
		} else {
			termbox.SetCell(x, y, c,
					fg.(termbox.Attribute),
					bg.(termbox.Attribute))
			x++
		}
	}
}

func TermPrintReverse(x   int,
		      y   int,
		      str FmtStr) {
	for i := len(str.str) - 1; i >= 0; i-- {
		c := rune(str.str[i])

		if c == '\t' {
			for j := 0; j < 4; j++ {
				termbox.SetCell(x, y, ' ',
						termbox.Attribute(str.clrs[0].clr.fg |
								  str.clrs[0].clr.attr),
						termbox.Attribute(str.clrs[0].clr.bg))
				x--
			}
		} else {
			termbox.SetCell(x, y, c,
					termbox.Attribute(str.clrs[0].clr.fg |
							  str.clrs[0].clr.attr),
					termbox.Attribute(str.clrs[0].clr.bg))
			x--
		}
	}
}

func TermPrintWidth(y   int,
		    str FmtStr) {
	w, _ := termbox.Size()
	x := 0

	var fg   = ColorDefault
	var bg   = ColorDefault
	var attr = AttrNone

	if str.tabbed {
		for j := 0; j < 4; j++ {
			termbox.SetCell(x, y, ' ',
					termbox.Attribute(fg | attr),
					termbox.Attribute(bg))
			x++
		}
	}

	cidx := -1
	if len(str.clrs) > 0 {
		cidx = 0
	}

	for i := 0; i < len(str.str); i++ {
		if cidx != -1 &&
		   i == str.clrs[cidx].idx {
			   fg   = str.clrs[cidx].clr.fg
			   bg   = str.clrs[cidx].clr.bg
			   attr = str.clrs[cidx].clr.attr
			   cidx++
			   if cidx == len(str.clrs) {
				   cidx = -1
			   }
		}

		termbox.SetCell(x, y, rune(str.str[i]),
				termbox.Attribute(fg | attr),
				termbox.Attribute(bg))
		x++
	}

	for x < w {
		termbox.SetCell(x, y, ' ',
				termbox.Attribute(fg | attr),
				termbox.Attribute(bg))
		x++
	}
}

func TermSetLineColorAttr(line int,
			  clr  ClrSpec) {
	w, _ := termbox.Size()

	cells := termbox.CellBuffer()

	for i := 0; i < w; i++ {
		cell := (line * w) + i
		//cells[cell].Fg = fg.(termbox.Attribute)
		cells[cell].Fg = termbox.Attribute(clr.fg | clr.attr)
		cells[cell].Bg = termbox.Attribute(clr.bg)
	}
}

func TermSetLineAttr(line int,
		     attr interface{}) {
	w, _ := termbox.Size()

	cells := termbox.CellBuffer()

	for i := 0; i < w; i++ {
		cell := (line * w) + i
		cells[cell].Fg = (cells[cell].Fg |
				  attr.(termbox.Attribute))
	}
}
func TermSetCursor(x int,
		   y int) {
	termbox.SetCursor(x, y)
}

func TermHideCursor() {
	termbox.HideCursor()
}

func TermSaveCellBuffer() []interface{} {
	cells := termbox.CellBuffer()
	copy_cells := make([]interface{}, len(cells))

	for i, c := range cells {
		copy_cells[i] = c
	}

	return copy_cells
}

func TermRestoreCellBuffer(cells []interface{}) {
	w, _ := termbox.Size()

	for i, c := range cells {
		termbox.SetCell((i % w), (i / w),
				c.(termbox.Cell).Ch,
				c.(termbox.Cell).Fg,
				c.(termbox.Cell).Bg)
	}
}

func TermSetCell(x  int,
		 y  int,
		 c  interface{}) {
	termbox.SetCell(x, y,
			c.(termbox.Cell).Ch,
			c.(termbox.Cell).Fg,
			c.(termbox.Cell).Bg)
}

func TermPollEvent() interface{} {
	return termbox.PollEvent()
}

func TermEventKeyCmp(ev  interface{},
		     key interface{}) bool {
	if ev.(termbox.Event).Type != termbox.EventKey {
		return false
	}

	switch key.(type) {
	case termbox.Key:
		return ev.(termbox.Event).Key == key.(termbox.Key)
	default:
		return ev.(termbox.Event).Ch == key
	}
}

func TermIsEventError(ev interface{}) bool {
	return ev.(termbox.Event).Type == termbox.EventError
}

func TermEventError(ev interface{}) error {
	return ev.(termbox.Event).Err
}

func TermInput(x     int,
	       y     int,
	       prime string) (string, error) {
	fg := termbox.ColorDefault
	bg := termbox.ColorDefault
	input := ""
	xx := x

	if prime != "" {
		for _, r := range prime {
			input += string(r)
			termbox.SetCell(xx, y, r, fg, bg)
			xx++
		}

		TermDrawScreen()
	}

mainloop:
	for {

	switch ev := termbox.PollEvent(); ev.Type {
	case termbox.EventError:
		log.Fatal(ev.Err)
	case termbox.EventKey:
		switch ev.Key {
		case termbox.KeyEsc, termbox.KeyCtrlC:
			return "", errors.New("cancelled")
		case termbox.KeyEnter:
			break mainloop
		case termbox.KeyCtrlU:
			for i := x; i < xx; i++ {
				termbox.SetCell(i, y, ' ', fg, bg)
			}
			input = ""
			xx = x
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if xx > x {
				xx--
				termbox.SetCell(xx, y, ' ', fg, bg)
				input = input[:(len(input)-1)]
			}
		case termbox.KeySpace:
			input += string(' ')
			termbox.SetCell(xx, y, ' ', fg, bg)
			xx++
		case termbox.KeyTab:
			input += string('\t')
			termbox.SetCell(xx, y, '\t', fg, bg)
			xx++
		default:
			if ev.Ch != 0 {
				input += string(ev.Ch)
				termbox.SetCell(xx, y, ev.Ch, fg, bg)
				xx++
			}
		}

		TermDrawScreen()
	}

	} /* for */

	return input, nil
}

func TermTitle(row   int,
	       left  string,
	       right string) {
	w, _ := TermSize()

	TermPrintWidth(row, ColorFmt(left, TitleFg, TitleBg, TitleAttr))
	TermPrintReverse((w - 1), row, ColorFmt(right, TitleFg, TitleBg, TitleAttr))
}

