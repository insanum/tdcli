
package main

import (
	"net/url"
	//"fmt"
	"strconv"
	"encoding/json"
	"sort"
	"strings"
	"os"
	"log"
	"regexp"
	"time"
)

var NoteListTitle string = "TOODLEDO: NOTES"

var NoteTemplate string =
`Title: <title>
Folder:
---
<note text>
`

const (
	NotesSortedAlpha = iota
	NotesSortedModified
	NotesSortedSearch
)

type TDNoteList struct {
	notes           []TDNote
	sorted          int
	sbLastSelectIdx int
	sb              ScrollBuffer
	sbLines         []ScrollLine
}

var TDNotes TDNoteList

/*
 * Send a new note to Toodledo servers.
 */
func TDNotesAddNote(title  string,
		    folder string,
		    text   string) {
	var notes []TDNote

	data := map[string]string{
		    "title":  title,
		    "folder": strconv.FormatInt(TDFolders.NameToId(folder), 10),
		    "added":  strconv.FormatInt(time.Now().UTC().Unix(), 10),
		    "text":   text }

	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	v := url.Values{}
	v.Set("notes", string(b))

	TDGetData("notes/add.php", v, &notes)

	if  notes[0].ErrorCode != 0 {
		StatusMessage(notes[0].ErrorDesc)
		return
	}

	TDNotes.notes = append(TDNotes.notes, notes[0])
	StatusMessage("Note added")
	TDFileCacheWrite()
	TDNotes.NoteListSortModified(true)
}

/*
 * Delete a note from Toodledo servers.
 */
func TDNotesDeleteNote(idx int) {
	var notes []TDNote

	data := []string{ strconv.FormatInt(TDNotes.notes[idx].ID, 10) }
	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	v := url.Values{}
	v.Set("notes", string(b))

	TDGetData("notes/delete.php", v, &notes)

	if  notes[0].ErrorCode != 0 {
		StatusMessage(notes[0].ErrorDesc)
		return
	}

	if notes[0].ID == TDNotes.notes[idx].ID {
		TDNotes.notes = append(TDNotes.notes[:idx],
				       TDNotes.notes[idx+1:]...)
		StatusMessage("Note deleted")
		TDFileCacheWrite()
		TDNotes.NoteListSortModified(true)
	} else {
		StatusMessage("ERROR: Unknown response")
	}
}

/*
 * Sort note list alphabetically by title.
 */
type NoteListSortByAlpha []TDNote
func (notes NoteListSortByAlpha) Len() int {
	return len(notes)
}
func (notes NoteListSortByAlpha) Swap(i int,
				      j int) {
	notes[i], notes[j] = notes[j], notes[i]
}
func (notes NoteListSortByAlpha) Less(i int,
				      j int) bool {
	return notes[i].Title < notes[j].Title
}

/*
 * Sort note list by last modified date.
 */
type NoteListSortByModified []TDNote
func (notes NoteListSortByModified) Len() int {
	return len(notes)
}
func (notes NoteListSortByModified) Swap(i int,
					 j int) {
	notes[i], notes[j] = notes[j], notes[i]
}
func (notes NoteListSortByModified) Less(i int,
					 j int) bool {
	return notes[i].Modified > notes[j].Modified
}


func (notes *TDNoteList) NoteListSetIndexes() {
	for i := 0; i < len(notes.notes); i++ {
		notes.notes[i].idx = i
	}
}

func (notes *TDNoteList) NoteListSortAlpha(draw bool) {
	sort.Sort(NoteListSortByAlpha(notes.notes))
	notes.sorted = NotesSortedAlpha
	notes.NoteListSetIndexes()

	notes.sbLines = make([]ScrollLine, len(notes.notes))
	notes.sbLastSelectIdx = 0

	for i := 0; i < len(notes.notes); i++ {
		notes.sbLines[i].text = notes.notes[i].Format()
		notes.sbLines[i].selectable = true
		notes.sbLines[i].realIdx = notes.notes[i].idx
	}

	notes.sb.Init(notes.NumHeaderLines(),
		      StatusLines(),
		      &notes.sbLines)

	if draw {
		notes.DrawPage()
	}
}

func (notes *TDNoteList) NoteListSortModified(draw bool) {
	sort.Sort(NoteListSortByModified(notes.notes))
	notes.sorted = NotesSortedModified
	notes.NoteListSetIndexes()

	notes.sbLines = make([]ScrollLine, len(notes.notes))
	notes.sbLastSelectIdx = 0

	for i := 0; i < len(notes.notes); i++ {
		notes.sbLines[i].text = notes.notes[i].Format()
		notes.sbLines[i].selectable = true
		notes.sbLines[i].realIdx = notes.notes[i].idx
	}

	notes.sb.Init(notes.NumHeaderLines(),
		      StatusLines(),
		      &notes.sbLines)

	if draw {
		notes.DrawPage()
	}
}

func (notes *TDNoteList) SearchNotes(draw bool) {
	search, err := AskString("Search:", "")
	if err != nil {
		return
	}

	if search == "" {
		return
	}

	rxp, err := regexp.Compile(`(?i)` + search)
	if err != nil {
		log.Fatal(err)
	}

	/* pre-sort by modified */
	sort.Sort(NoteListSortByModified(notes.notes))
	notes.sorted = NotesSortedSearch

	notes.sbLines = []ScrollLine{}
	notes.sbLastSelectIdx = 0

	for i := 0; i < len(notes.notes); i++ {
		if rxp.FindString(notes.notes[i].Text)                       != "" ||
		   rxp.FindString(notes.notes[i].Title)                      != "" ||
		   rxp.FindString(TDFolders.IdToName(notes.notes[i].Folder)) != "" {
			var line ScrollLine
			line.text = notes.notes[i].Format()
			line.selectable = true
			line.realIdx = notes.notes[i].idx
			notes.sbLines = append(notes.sbLines, line)
		}
	}

	notes.sb.Init(notes.NumHeaderLines(),
		      StatusLines(),
		      &notes.sbLines)

	if draw {
		notes.DrawPage()
	}
}

func (notes *TDNoteList) NumHeaderLines() int {
	return 1
}

func (notes *TDNoteList) NumBodyLines() int {
	return len(notes.sbLines)
}

func (notes *TDNoteList) MoveCursorDown() {
	notes.sb.CursorDown()
}

func (notes *TDNoteList) MoveCursorUp() {
	notes.sb.CursorUp()
}

func (notes *TDNoteList) ScrollBodyDown(lines int) {
	notes.sb.ScrollDown(lines)
}

func (notes *TDNoteList) ScrollBodyUp(lines int) {
	notes.sb.ScrollUp(lines)
}

func (notes *TDNoteList) UpdateHeader() {
	str := IdxString(((notes.sb.row - notes.NumHeaderLines()) + 1),
			 notes.NumBodyLines())
	TermTitle(0, NoteListTitle, str)

	TermDrawScreen()
}

func (notes *TDNoteList) DrawPage() {
	TermClearScreen(false)

	str := IdxString(1, notes.NumBodyLines())
	TermTitle(0, NoteListTitle, str)

	notes.sb.Draw()
}

func (notes *TDNoteList) NextPage() interface{} {
	return nil
}

func (notes *TDNoteList) PrevPage() interface{} {
	return nil
}

func (notes *TDNoteList) ChildPage() interface{} {
	idx := notes.sb.CursorToIdx()
	if notes.sbLines[idx].selectable {
		notes.sbLastSelectIdx = idx
		return &notes.notes[notes.sbLines[idx].realIdx]
	} else {
		return nil
	}
}

func (notes *TDNoteList) ParentPage() interface{} {
	return nil
}

func (notes *TDNoteList) AddNote() {
	editor := os.ExpandEnv("$EDITOR")
	if editor == "" {
		StatusMessage("$EDITOR not defined")
		return
	}

	fname := TempFileWrite(NoteTemplate)

	ExecCommand(editor, fname)

	note := TempFileRead(fname)
	TempFileDelete(fname)

	if strings.TrimSpace(note) ==
	   strings.TrimSpace(NoteTemplate) {
		/* nothing changed */
		StatusMessage("Cancelled")
		return
	}

	rxpTitle, err := regexp.Compile(`^\s*Title:(.*)$`)
	if err != nil {
		log.Fatal(err)
	}
	rxpFolder, err := regexp.Compile(`^\s*Folder:(.*)$`)
	if err != nil {
		log.Fatal(err)
	}
	rxpSep, err := regexp.Compile(`^\s*---\s*$`)
	if err != nil {
		log.Fatal(err)
	}

	splnote := strings.SplitN(note, "\n", 4)
	var res []string

	res = rxpTitle.FindStringSubmatch(splnote[0])
	if res == nil {
		StatusMessage("ERROR: Invalid note title")
		return
	}

	ntitle := strings.TrimSpace(res[1])

	res = rxpFolder.FindStringSubmatch(splnote[1])
	if res == nil {
		StatusMessage("ERROR: Invalid note folder")
		return
	}

	nfolder := strings.TrimSpace(res[1])
	/* XXX verify folder name (ignore or re-edit?) */

	res = rxpSep.FindStringSubmatch(splnote[2])
	if res == nil {
		StatusMessage("ERROR: Invalid note format")
		return
	}

	ntext := strings.TrimSpace(splnote[3])

	TDNotesAddNote(ntitle, nfolder, ntext)
	notes.DrawPage()
}

func (notes *TDNoteList) DeleteNote() {
	if !AskYesNo("Delete note [y/n]?:") {
		return
	}

	TDNotesDeleteNote(notes.sbLines[notes.sb.CursorToIdx()].realIdx)
	notes.DrawPage()
}

/*
 * Sync all notes from Toodledo servers.
 */
func TDNotesSync(full_sync bool) {
	var after int64
	var v url.Values
	var notes []TDNote
	var dnotes []TDNoteDeleted

	/*
	 * Get any new or edited notes...
	 */

	if full_sync {
		after = 0
	} else {
		if tda.AccountInfo.LastEditNote >= tdx.LastEditNote {
			goto sync_notes_deleted
		}

		after = tda.AccountInfo.LastEditNote
	}

	v = url.Values{}
	v.Set("after", strconv.FormatInt(after, 10))

	TDGetData("notes/get.php", v, &notes)

	//fmt.Printf("num=%d total=%d\n", notes[0].Num, notes[0].Total)

	if full_sync {
		TDNotes.notes = notes[1:]
		return
	}

	for _, n1 := range notes[1:] {
		updated := false
		for i, n2 := range TDNotes.notes {
			if n1.ID == n2.ID {
				//fmt.Printf("Updating Note -> %s (%d)\n", n1.Title, n1.ID)
				TDNotes.notes[i] = n1
				updated = true
			}
		}
		if !updated {
			//fmt.Printf("Adding Note -> %s (%d)\n", n1.Title, n1.ID)
			TDNotes.notes = append(TDNotes.notes, n1)
		}
	}

	/*
	 * Get any deleted notes...
	 */

sync_notes_deleted:

	if tda.AccountInfo.LastDeleteNote >= tdx.LastDeleteNote {
		return
	}

	after = tda.AccountInfo.LastDeleteNote

	v = url.Values{}
	v.Set("after", strconv.FormatInt(after, 10))

	TDGetData("notes/deleted.php", v, &dnotes)

	//fmt.Printf("num=%d\n", dnotes[0].Num)

	for _, dn1 := range dnotes[1:] {
		for i, dn2 := range TDNotes.notes {
			if dn1.ID == dn2.ID {
				//fmt.Printf("Deleting Note -> %s (%d)\n", dn2.Title, dn2.ID)
				TDNotes.notes = append(TDNotes.notes[:i],
						       TDNotes.notes[i+1:]...)
			}
		}
	}
}

