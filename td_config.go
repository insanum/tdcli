
package main

import (
	"os"
)

var (

OauthHttpRedirectUrl  string = "http://127.0.0.1"
OauthHttpRedirectPort int    = 56789

ToodledoUrl     string = "https://api.toodledo.com/3/"
CacheFileName   string = os.ExpandEnv("$HOME") + "/.toodledo_cache.json"
AccountFileName string = os.ExpandEnv("$HOME") + "/.toodledo_account.json"
LogFileName     string = os.ExpandEnv("$HOME") + "/.toodledo_tdcli.log"

/*
 * Colors Numbers (1-based):
 *       0: default
 *     1-8:        black, red, green, yellow, blue, magenta, cyan, white
 *    9-16: (BOLD) black, red, green, yellow, blue, magenta, cyan, white
 *  17-232: 216 colors 
 * 233-256: 24 shades of grey
 *
 * Foreground/Background color:
 *   <num> = Color
 *   0     = Reset/Default
 *
 * Attribute:
 *   B = Bold
 *   U = Underline
 *   R = Reverse
 *   0 = Reset/None
 *
 * Set Foreground/Background/Attribute
 *   {#<foreground[num]>,<background[num]>,<attribute[BUR0]>}
 *
 * Reset Foreground/Background/Attribute
 *   {#-}
 */

/*
 * Task formats:
 * {i}  - ID
 * {t}  - Title
 * {*}  - Star
 * {p}  - Priority
 * {f}  - Folder
 * {l}  - Location
 * {c}  - Context
 * {g}  - Goal
 * {pi} - Parent ID
 * {pt} - Parent Title
 * {nc} - Number of Children
 * {dd} - Due Date
 * {dt} - Due Time
 * {dm} - Due Date Mod
 * {sd} - Start Date
 * {st} - Start Time
 * {r}  - Remind
 * {rp} - Repeat
 * {s}  - Status
 * {lg} - Length
 * {md} - Modified Date/Time
 * {cd} - Completed Date/Time
 * {n}  - Note Text
 */
TaskFmtHeader string = "{#46,236,B}%{t}{#-}"
TaskFmt       string = "%1.1{*} {#161,0,0}%-10.10{f}{#-} %1.1{dm}%-10.10{dd} %-5.5{dt} {#77,0,0}%-11.11{p}{#-} %{t}"

/*
 * Note formats:
 * {i}  - ID
 * {t}  - Title
 * {f}  - Folder
 * {md} - Modified Date/Time
 * {ad} - Added Date/Time
 * {n}  - Note Text
 */
NoteFmt string = "{#161,0,0}%-10.10{f}{#-} %{t}"

/*
 * Help formats:
 * {t} - Help group title
 * {n} - Keybind name
 * {k} - Keybind shortcut
 * {d} - Keybind description
 */
HelpFmtHeader  string = "{#46,0,0}%{t}{#-}"
HelpFmtKeybind string = "{#173,0,0}%20{n}{#-}  {#41,0,B}%-5{k}{#-}  %{d}"

TitleFg   int = 17
TitleBg   int = 209
TitleAttr int = AttrBold

CursorFg   int = ColorDefault
CursorBg   int = 54
CursorAttr int = AttrBold

StatusFg   int = 1
StatusBg   int = 226
StatusAttr int = AttrNone

SortOrder1 = [...]string{ "due",      "star",     "priority"}
SortOrder2 = [...]string{ "folder",   "due",      "priority"}
SortOrder3 = [...]string{ "priority", "due",      "star"}
SortOrder4 = [...]string{ "star",     "priority", "due"}

Keybinds = map[string]struct{
	key   string
	descr string
}{
	/* general */
	"quit":      { "Q",     "Quit" },
	"enter":     { "Enter", "Select current highlighted entry" },
	"back":      { "q",     "Go back to parent list" },
	"down":      { "j",     "Move down to next entry" },
	"up":        { "k",     "Move up to previous entry" },
	"sdown":     { "CTL-E", "Scroll down one entry" },
	"sup":       { "CTL-Y", "Scroll up one entry" },
	"sdownh":    { "CTL-D", "Scroll down half page" },
	"suph":      { "CTL-U", "Scroll up half page" },
	"sdownf":    { "Space", "Scroll down full page" },
	"supf":      { "b",     "Scroll up full page" },
	"notes":     { "n",     "View note list" },
	"tasks":     { "t",     "View task list" },
	"folders":   { "f",     "View folder list" },
	"locations": { "l",     "View location list" },
	"contexts":  { "c",     "View context list" },
	"goals":     { "g",     "View goal list" },
	"next":      { "J",     "Next (Task/Note)" },
	"previous":  { "K",     "Previous (Task/Note)" },
	"search":    { "/",     "Search (Tasks/Notes)" },

	/* task list */
	"tl_sort1":    { "1", "Sort order 1" },
	"tl_sort2":    { "2", "Sort order 2" },
	"tl_sort3":    { "3", "Sort order 3" },
	"tl_sort4":    { "4", "Sort order 4" },
	"tl_add":      { "A", "Add a new task" },
	"tl_addc":     { "C", "Add a new child task" },
	"tl_delete":   { "D", "Delete a task" },
	"tl_complete": { "X", "Complete a task" },

	/* task */
	"t_toggle_star":   { "*", "Toggle task star" },
	"t_edit_title":    { "T", "Edit task title" },
	"t_edit_priority": { "P", "Edit task priority" },
	"t_edit_folder":   { "F", "Edit task folder" },
	"t_edit_tags":     { "S", "Edit task tags" },
	"t_edit_due":      { "D", "Edit task due date/time" },
	"t_edit_remind":   { "R", "Edit task reminder" },
	"t_edit_location": { "L", "Edit task location" },
	"t_edit_context":  { "C", "Edit task context" },
	"t_edit_goal":     { "G", "Edit task goal" },
	"t_edit_note":     { "N", "Edit task note" },

	/* note list */
	"nl_sort_alpha": { "a", "Sort notes alphabetically" },
	"nl_sort_mod":   { "m", "Sort notes last modified" },
	"nl_add_note":   { "A", "Add a new note" },
	"nl_del_note":   { "D", "Delete a note" },

	/* note */
	"n_edit_folder": { "F", "Edit note folder" },
	"n_edit_text":   { "E", "Edit note with $EDITOR" },
	"n_pager":       { "P", "View note with $PAGER" },

	/* UDLs */
	"udl_new":    { "A", "Add a new entry" },
	"udl_edit":   { "E", "Edit an entry" },
	"udl_delete": { "D", "Delete an entry" },
}

)

