
package main

import (
	//"fmt"
)

var KeybindsHelpOrder = []string{
	"General", /* header */
	"quit",
	"enter",
	"back",
	"down",
	"up",
	"sdown",
	"sup",
	"sdownh",
	"suph",
	"sdownf",
	"supf",
	"notes",
	"tasks",
	"folders",
	"locations",
	"contexts",
	"goals",
	"next",
	"previous",
	"search",
	"",
	"Task List", /* header */
	"tl_sort1",
	"tl_sort2",
	"tl_sort3",
	"tl_sort4",
	"tl_add",
	"tl_addc",
	"tl_delete",
	"tl_complete",
	"",
	"Task", /* header */
	"t_toggle_star",
	"t_edit_title",
	"t_edit_priority",
	"t_edit_folder",
	"t_edit_tags",
	"t_edit_due",
	"t_edit_remind",
	"t_edit_location",
	"t_edit_context",
	"t_edit_goal",
	"t_edit_note",
	"",
	"Note List", /* header */
	"nl_sort_alpha",
	"nl_sort_mod",
	"nl_add_note",
	"nl_del_note",
	"",
	"Note", /* header */
	"n_edit_folder",
	"n_edit_text",
	"n_pager",
	"",
	"User-defined Lists", /* header */
	"udl_new",
	"udl_edit",
	"udl_delete",
}

var HelpTitle string = "TOODLEDO: HELP"

type TDHelpPage struct {
	sbLastSelectIdx int
	sb              ScrollBuffer
	sbLines         []ScrollLine
}

var TDHelp TDHelpPage

func (help *TDHelpPage) NumHeaderLines() int {
	return 1
}

func (help *TDHelpPage) NumBodyLines() int {
	return len(help.sbLines)
}

func (help *TDHelpPage) MoveCursorDown() {
	help.sb.CursorDown()
}

func (help *TDHelpPage) MoveCursorUp() {
	help.sb.CursorUp()
}

func (help *TDHelpPage) ScrollBodyDown(lines int) {
	help.sb.ScrollDown(lines)
}

func (help *TDHelpPage) ScrollBodyUp(lines int) {
	help.sb.ScrollUp(lines)
}

func (help *TDHelpPage) UpdateHeader() {
	TermDrawScreen()
}

func (help *TDHelpPage) DrawPage() {
	help.sbLines = make([]ScrollLine, len(KeybindsHelpOrder))
	for i, name := range KeybindsHelpOrder {
		if name == "" {
			help.sbLines[i].text = EmptyFmt("")
		} else {
			if value, ok := Keybinds[name]; ok {
				help.sbLines[i].text = HelpKeybindFormat(name, value.key, value.descr)
			} else {
				help.sbLines[i].text = HelpHeaderFormat(name)
			}
		}

		help.sbLines[i].selectable = true
		help.sbLines[i].realIdx = i
	}

	help.sb.Init(help.NumHeaderLines(),
		     StatusLines(),
		     &help.sbLines)

	TermClearScreen(false)

	TermTitle(0, HelpTitle, "")

	help.sb.Draw()
}

func (help *TDHelpPage) NextPage() interface{} {
	return nil
}

func (help *TDHelpPage) PrevPage() interface{} {
	return nil
}

func (help *TDHelpPage) ChildPage() interface{} {
	return nil
}

func (help *TDHelpPage) ParentPage() interface{} {
	return nil
}

