
package td

import (
	"net/url"
	"fmt"
	"strconv"
	"log"
	"time"
	"strings"
	"os"
	"encoding/json"
)

var NoteTitle string = "TOODLEDO: NOTE"

type TDNote struct {
	Num           int64  `json:"num"`       /* element from 'notes/get.php. */
	Total         int64  `json:"total"`     /* element from 'notes/get.php' */
	ErrorCode     int64  `json:"errorCode"` /* element from 'notes/edit.php' */
	ErrorDesc     string `json:"errorDesc"` /* element from 'notes/edit.php' */
	Ref           string `json:"ref"`       /* element from 'notes/edit.php' */

	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Folder        int64  `json:"folder"`
	Modified      int64  `json:"modified"`
	Added         int64  `json:"added"`
	Private       int64  `json:"private"`
	Text          string `json:"text"`

	idx           int
	sb            ScrollBuffer
	sbLines       []ScrollLine
}

type TDNoteDeleted struct {
	Num   int64 `json:"num"` /* first element 'notes/deleted.php. */
	ID    int64 `json:"id"`
	Stamp int64 `json:"stamp"`
}

/*
 * Sync a note back to Toodledo servers.
 */
func TDNoteSyncNote(idx  int,
		    data map[string]string) {
	var notes []TDNote

	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	v := url.Values{}
	v.Set("notes", string(b))

	TDGetData("notes/edit.php", v, &notes)

	if  notes[0].ErrorCode != 0 {
		StatusMessage(notes[0].ErrorDesc)
		return
	}

	if notes[0].ID == TDNotes.notes[idx].ID {
		TDNotes.notes[idx]     = notes[0]
		TDNotes.notes[idx].idx = idx
		StatusMessage("Note updated")
		TDFileCacheWrite()
	} else {
		StatusMessage("ERROR: Unknown response")
	}
}

func (note *TDNote) NoteSyncText() {
	data := map[string]string{
		    "id":   strconv.FormatInt(note.ID, 10),
		    "text": note.Text }
	TDNoteSyncNote(note.idx, data)
}

func (note *TDNote) NoteSyncFolder() {
	data := map[string]string{
		    "id":     strconv.FormatInt(note.ID, 10),
		    "folder": strconv.FormatInt(note.Folder, 10) }
	TDNoteSyncNote(note.idx, data)
}

func (note *TDNote) NumHeaderLines() int {
	return 2
}

func (note *TDNote) NumBodyLines() int {
	return len(note.sbLines)
}

func (note *TDNote) MoveCursorDown() {
	note.sb.CursorDown()
}

func (note *TDNote) MoveCursorUp() {
	note.sb.CursorUp()
}

func (note *TDNote) ScrollBodyDown(lines int) {
	note.sb.ScrollDown(lines)
}

func (note *TDNote) ScrollBodyUp(lines int) {
	note.sb.ScrollUp(lines)
}

func (note *TDNote) UpdateHeader() {
	TermDrawScreen()
}

func (note *TDNote) DrawPage() {
	lines := strings.Split(note.Text, "\n")
	note.sbLines = make([]ScrollLine, len(lines))
	for i := 0; i < len(lines); i++ {
		note.sbLines[i].text = EmptyFmt(lines[i])
		note.sbLines[i].selectable = true
		note.sbLines[i].realIdx = i
	}

	note.sb.Init(note.NumHeaderLines(),
		     StatusLines(),
		     &note.sbLines)

	TermClearScreen(false)

	TermTitle(0, NoteTitle, "")

	strL := fmt.Sprintf("(%s) %s", TDFolders.IdToName(note.Folder), note.Title)
	strR := time.Unix(note.Modified, 0).Format("2006/1/2 15:04")
	TermTitle(1, strL, strR)

	note.sb.Draw()
}

func (note *TDNote) NextPage() interface{} {
	if TDNotes.sbLastSelectIdx < (len(TDNotes.sbLines) - 1) {
		TDNotes.sb.ScrollUp(1)
		TDNotes.sbLastSelectIdx++
		return &TDNotes.notes[TDNotes.sbLines[TDNotes.sbLastSelectIdx].realIdx]
	}
	return nil
}

func (note *TDNote) PrevPage() interface{} {
	if TDNotes.sbLastSelectIdx > 0 {
		TDNotes.sb.ScrollDown(1)
		TDNotes.sbLastSelectIdx--
		return &TDNotes.notes[TDNotes.sbLines[TDNotes.sbLastSelectIdx].realIdx]
	}
	return nil
}

func (note *TDNote) ChildPage() interface{} {
	return nil
}

func (note *TDNote) ParentPage() interface{} {
	return &TDNotes
}

func (note *TDNote) EditNoteText() {
	editor := os.ExpandEnv("$EDITOR")
	if editor == "" {
		StatusMessage("$EDITOR not defined")
		return
	}

	fname := TempFileWrite(note.Text)

	ExecCommand(editor, fname)

	newText := TempFileRead(fname)
	TempFileDelete(fname)

	if strings.TrimSpace(newText) ==
	   strings.TrimSpace(note.Text) {
		/* nothing changed */
		StatusMessage("Cancelled")
		return
	}

	note.Text = newText

	note.NoteSyncText()
	note.DrawPage()
}

func (note *TDNote) EditNoteFolder() {
	folder, err := AskString("Folder name:", "")
	if err != nil {
		return
	}

	if folder == "" {
		if !AskYesNo("Clear folder [y/n]?:") {
			return
		}

		folder = "none"
	}

	if folder == "none" {
		note.Folder = 0
	} else {
		id := TDFolders.NameToId(folder)
		if id == 0 {
			StatusMessage("ERROR: Invalid folder name")
			return
		}

		note.Folder = id
	}

	note.NoteSyncFolder()
	note.DrawPage()
}

func (note *TDNote) PagerNote() {
	pager := os.ExpandEnv("$PAGER")
	if pager == "" {
		StatusMessage("$EDITOR not defined")
		return
	}

	fname := TempFileWrite(note.Text)

	ExecCommand(pager, fname)

	TempFileDelete(fname)
}

