
package td

type CursorPosition struct {
	row    int
	column int
}

type ScrollLine struct {
	text       FmtStr
	selectable bool
	realIdx    int
}

type ScrollBuffer struct {
	CursorPosition
	firstLine      int // first 'lines' index shown in scroll buffer
	headerLines    int // number of lines before scroll buffer
	footerLines    int // number of lines after scroll buffer
	bufferLines    int // number of lines used by scroll buffer
	totalLines     int // total number of 'lines' to scroll
	lines          *[]ScrollLine // data lines
}

func (sb *ScrollBuffer) Init(hl int, /* header lines */
			     fl int, /* footer lines */
			     l  *[]ScrollLine) {
	_, h := TermSize()
	sb.firstLine   = 0
	sb.headerLines = hl
	sb.footerLines = fl
	sb.bufferLines = (h - (hl + fl))
	sb.totalLines  = len(*l)
	sb.lines       = l

	//sb.SetCursor(hl, 0, true)
	sb.row    = hl
	sb.column = 0
}

func (sb *ScrollBuffer) SetCursorAttr() {
	TermSetLineColorAttr(sb.row, ClrSpec{CursorFg, CursorBg, CursorAttr})
}

func (sb *ScrollBuffer) Draw() {
	cnt := sb.bufferLines
	if (sb.totalLines - sb.firstLine) < sb.bufferLines {
		cnt = (sb.totalLines - sb.firstLine)
	}

	if cnt == 0 {
		TermDrawScreen()
		return
	}

	for i := 0; i < cnt; i++ {
		TermPrintWidth((i + sb.headerLines), (*sb.lines)[i + sb.firstLine].text)
	}

	if (*sb.lines)[sb.CursorToIdx()].selectable {
		/* highlight cursor line */
		sb.SetCursorAttr()
	}

	TermDrawScreen()
}

func (sb *ScrollBuffer) ScrollDown(numLines int) {
	if (sb.firstLine == 0) ||
	   (sb.totalLines <= sb.bufferLines) {
		return
	}

	sb.firstLine -= numLines
	if sb.firstLine < 0 {
		sb.firstLine = 0
	}

	sb.Draw()
}

func (sb *ScrollBuffer) ScrollUp(numLines int) {
	if (sb.totalLines <= sb.bufferLines) ||
	   ((sb.totalLines - sb.firstLine) <= sb.bufferLines) {
		return
	}

	sb.firstLine += numLines
	if (sb.totalLines - sb.firstLine) <= sb.bufferLines {
		sb.firstLine = (sb.totalLines - sb.bufferLines)
	}

	sb.Draw()
}

func (sb *ScrollBuffer) CursorToIdx() int {
	return (sb.row + sb.firstLine - sb.headerLines)
}

func (sb *ScrollBuffer) SetCursor(row      int,
				  column   int,
				  absolute bool) {
	if absolute {
		sb.row    = row
		sb.column = column
	} else {
		sb.row    += row
		sb.column += column
	}

	TermSetCursor(sb.row, sb.column)
	TermHideCursor()
}

func (sb *ScrollBuffer) MoveCursor(dir int) {
	var sl ScrollLine

	new_line := (sb.row + dir)

	sl = (*sb.lines)[sb.CursorToIdx()]
	if sl.selectable {
		TermPrintWidth(sb.row, sl.text)
	}

	sb.SetCursor(new_line, sb.column, true)

	sl = (*sb.lines)[sb.CursorToIdx()]
	if sl.selectable {
		sb.SetCursorAttr()
	}

	state.page.UpdateHeader()
}

func (sb *ScrollBuffer) CursorDown() {
	_, h := TermSize()

	if sb.row == (h - sb.footerLines - 1) {
		sb.ScrollUp(1)
		return
	}

	if sb.totalLines < sb.bufferLines {
		if sb.row < (sb.totalLines + sb.headerLines - 1) {
			sb.MoveCursor(1)
		}
	} else {
		sb.MoveCursor(1)
	}
}

func (sb *ScrollBuffer) CursorUp() {
	if sb.row == sb.headerLines {
		sb.ScrollDown(1)
	} else {
		sb.MoveCursor(-1)
	}
}

